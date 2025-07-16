package sales

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/finance"
	"github.com/AMETORY/ametory-erp-modules/inventory"
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/google/uuid"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SalesService struct {
	ctx              *context.ERPContext
	db               *gorm.DB
	financeService   *finance.FinanceService
	inventoryService *inventory.InventoryService
}

// Migrate applies database schema changes for the sales module.
// It automates the migration of sales-related models, ensuring that the database schema
// is up to date with the current definitions of SalesModel, SalesItemModel, and SalesPaymentModel.
// If successful, it returns nil; otherwise, it returns an error indicating what went wrong.
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&models.SalesModel{}, &models.SalesItemModel{}, &models.SalesPaymentModel{})
}

// NewSalesService creates a new instance of SalesService with the given database connection, context, finance service and inventory service.
func NewSalesService(db *gorm.DB, ctx *context.ERPContext, financeService *finance.FinanceService, inventoryService *inventory.InventoryService) *SalesService {
	return &SalesService{db: db, ctx: ctx, financeService: financeService, inventoryService: inventoryService}
}

// CreateSales creates a new sales document in the database and performs relevant accounting entries.
// If the sales document has items with a sale account and/or an asset account, transactions will be created
// for the sale and the asset account. If the sales document has a payment account, the sales document will be
// marked as paid. If the sales document has a partial payment, the sales document will be marked as partial.
// If successful, it returns nil; otherwise, it returns an error indicating what went wrong.
func (s *SalesService) CreateSales(data *models.SalesModel) error {
	var companyID *string
	if s.ctx.Request.Header.Get("ID-Company") != "" {
		compID := s.ctx.Request.Header.Get("ID-Company")
		companyID = &compID
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		data.CompanyID = companyID
		if err := s.db.Create(data).Error; err != nil {
			tx.Rollback()
			return err
		}
		if s.financeService.TransactionService != nil {
			paid := 0.0
			for _, v := range data.Items {
				if v.SaleAccountID != nil {
					s.financeService.TransactionService.CreateTransaction(&models.TransactionModel{
						Date:               data.SalesDate,
						AccountID:          v.SaleAccountID,
						Description:        "Penjualan " + data.SalesNumber,
						Notes:              data.Description,
						TransactionRefID:   &data.ID,
						TransactionRefType: "sales",
						CompanyID:          companyID,
					}, v.Total)
				}
				if v.AssetAccountID != nil {
					s.financeService.TransactionService.CreateTransaction(&models.TransactionModel{
						Date:               data.SalesDate,
						AccountID:          v.AssetAccountID,
						Description:        "Penjualan " + data.SalesNumber,
						Notes:              data.Description,
						TransactionRefID:   &data.ID,
						TransactionRefType: "sales",
						CompanyID:          companyID,
					}, v.Total)
					acc, err := s.financeService.AccountService.GetAccountByID(*v.AssetAccountID)
					if err != nil {
						return err
					}
					if acc.Type == models.ASSET {
						paid += v.Total
					}
				}

			}
			if paid > 0 {
				data.Paid = paid
				if err := tx.Save(data).Error; err != nil {
					tx.Rollback()
					return err
				}
			}

			if paid < data.Total {
				data.Status = "partial"
				if err := tx.Save(data).Error; err != nil {
					tx.Rollback()
					return err
				}
			}
		}

		return nil
	})

}

// CreatePayment creates a new payment transaction associated with a sales order.
//
// The function takes as parameters:
//   - salesID: the ID of the sales order to be paid
//   - date: the date of the payment
//   - amount: the amount of the payment
//   - accountReceivableID: the ID of the account receivable associated with the sales order
//   - accountAssetID: the ID of the asset account to be debited
//
// It performs the following operations:
//  1. Retrieves the sales order from the database and checks if its status is "pending".
//  2. Checks if the payment amount is greater than the remaining balance of the sales order. If so, it returns an error.
//  3. Creates a new transaction record in the database with the following details:
//     - Date: the provided date
//     - AccountID: the ID of the asset account to be debited
//     - Description: "Pembayaran [sales number]"
//     - Notes: the description of the sales order
//     - TransactionRefID: the ID of the sales order
//     - TransactionRefType: "sales"
//     - CompanyID: the ID of the company associated with the sales order
//     - Debit: the payment amount
//  4. If accountReceivableID is not nil, it creates another transaction record with the following details:
//     - AccountID: the ID of the account receivable associated with the sales order
//     - Credit: the payment amount
//  5. Updates the sales order record in the database with the new paid amount.
//  6. If the paid amount is equal to the total amount, it updates the status of the sales order to "paid".
//  7. Commits the transaction if all operations are successful. Otherwise, it rolls back the transaction.
//
// Returns an error if any of the operations fail.
func (s *SalesService) CreatePayment(salesID string, date time.Time, amount float64, accountReceivableID *string, accountAssetID string) error {
	var companyID *string
	if s.ctx.Request.Header.Get("ID-Company") != "" {
		compID := s.ctx.Request.Header.Get("ID-Company")
		companyID = &compID
	}
	return s.db.Transaction(func(tx *gorm.DB) error {

		var data models.SalesModel
		if err := tx.Where("id = ?", salesID).First(&data).Error; err != nil {
			return err
		}

		if data.Paid+amount > data.Total {
			return errors.New("amount is greater than total")
		}

		if err := s.financeService.TransactionService.CreateTransaction(&models.TransactionModel{
			Date:               date,
			AccountID:          &accountAssetID,
			Description:        "Pembayaran " + data.SalesNumber,
			Notes:              data.Description,
			TransactionRefID:   &data.ID,
			TransactionRefType: "sales",
			CompanyID:          companyID,
		}, amount); err != nil {
			return err
		}
		if accountReceivableID != nil {
			if err := s.financeService.TransactionService.CreateTransaction(&models.TransactionModel{
				Date:               date,
				AccountID:          accountReceivableID,
				Description:        "Pembayaran " + data.SalesNumber,
				Notes:              data.Description,
				TransactionRefID:   &data.ID,
				TransactionRefType: "sales",
				CompanyID:          companyID,
			}, -amount); err != nil {
				return err
			}
		}

		data.Paid += amount
		if err := tx.Save(data).Error; err != nil {
			return err
		}

		if data.Paid == data.Total {
			data.Status = "paid"
			if err := tx.Save(data).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// UpdateSales updates the sales order data with the given ID.
//
// This function will update all fields of the sales order except for the ID.
//
// Returns an error if the update operation fails.
func (s *SalesService) UpdateSales(id string, data *models.SalesModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

func (s *SalesService) DeleteSales(id string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Where("transaction_ref_id = ? and transaction_secondary_ref_id = ?", id, id).Delete(&models.TransactionModel{}).Error
		if err != nil {
			return err
		}
		err = tx.Where("sales_id = ?", id).Delete(&models.SalesItemModel{}).Error
		if err != nil {
			return err
		}
		err = tx.Where("sales_id = ?", id).Delete(&models.SalesPaymentModel{}).Error
		if err != nil {
			return err
		}
		err = tx.Where("reference_id = ? or secondary_ref_id = ?", id, id).Delete(&models.StockMovementModel{}).Error
		if err != nil {
			return err
		}
		return tx.Where("id = ?", id).Delete(&models.SalesModel{}).Error
	})

}

// GetSalesByID retrieves a sales order from the database by ID.
//
// The function returns an error if the sales order is not found.
//
// The function also populates the Paid field of the sales order.
// If the sales order has a payment account and the type of the account is ASSET, the Paid field is set to the total amount of the sales order.
// Otherwise, the Paid field is the sum of the amounts of all payments associated with the sales order.
func (s *SalesService) GetSalesByID(id string) (*models.SalesModel, error) {
	var sales, refSales models.SalesModel
	err := s.db.Preload("Contact").Preload("SalesUser").Preload("PublishedBy", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, full_name")
	}).Preload("SalesPayments", func(db *gorm.DB) *gorm.DB {
		return db.Order("updated_at asc")
	}).Preload("PaymentAccount").Where("id = ?", id).First(&sales).Error
	if err != nil {
		return nil, err
	}
	if sales.RefID != nil {
		err = s.db.Where("id = ?", *sales.RefID).First(&refSales).Error
		if err != nil {
			return nil, err
		}
		sales.SalesRef = &refSales
	}
	paid := 0.0
	for _, v := range sales.SalesPayments {
		paid += v.Amount
	}

	// utils.LogJson(sales.PaymentAccount)
	if sales.PaymentAccount != nil {
		if sales.PaymentAccount.Type == "ASSET" {
			paid = sales.Total
		}
	}
	sales.Paid = paid
	s.db.Model(&sales).Where("id = ?", id).Update("paid", paid)
	return &sales, err
}

// GetSalesByCode retrieves a sales order from the database by code.
//
// The function returns an error if the sales order is not found.
func (s *SalesService) GetSalesByCode(code string) (*models.SalesModel, error) {
	var sales models.SalesModel
	err := s.db.Where("code = ?", code).First(&sales).Error
	return &sales, err
}

// GetSalesBySalesNumber retrieves a sales order from the database by sales number.
//
// The function returns an error if the sales order is not found.
func (s *SalesService) GetSalesBySalesNumber(salesNumber string) (*models.SalesModel, error) {
	var sales models.SalesModel
	err := s.db.Where("sales_number = ?", salesNumber).First(&sales).Error
	return &sales, err
}

// GetSales retrieves a paginated list of sales orders from the database.
//
// The function takes an http.Request and a search query string as input. The method
// preloads the sales user and contact of the sales order, and returns a pointer to a
// paginate.Page. The function utilizes pagination to manage the result set and applies any
// necessary request modifications using the utils.FixRequest utility.
//
// The function returns a paginated page of SalesModel and an error if the operation fails.
func (s *SalesService) GetSales(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.Preload("SalesUser").Preload("Contact")
	if search != "" {
		stmt = stmt.Where("sales.description ILIKE ? OR sales.code ILIKE ? OR sales.sales_number ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
			"%"+search+"%",
		)
	}
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	if request.URL.Query().Get("doc_type") != "" {
		stmt = stmt.Where("document_type = ?", request.URL.Query().Get("doc_type"))
	}
	if request.URL.Query().Get("status") != "" {
		stmt = stmt.Where("status = ?", request.URL.Query().Get("status"))
	}
	if request.URL.Query().Get("start_date") != "" {
		stmt = stmt.Where("sales_date >= ?", request.URL.Query().Get("start_date"))
	}
	if request.URL.Query().Get("end_date") != "" {
		stmt = stmt.Where("sales_date <= ?", request.URL.Query().Get("end_date"))
	}
	if request.URL.Query().Get("sales_user_id") != "" {
		stmt = stmt.Where("sales_user_id = ?", request.URL.Query().Get("sales_user_id"))
	}
	if request.URL.Query().Get("order") != "" {
		stmt = stmt.Order(request.URL.Query().Get("order"))
	}
	stmt = stmt.Model(&models.SalesModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.SalesModel{})
	page.Page = page.Page + 1
	return page, nil
}

// UpdateStock updates the stock of a sales order and its items.
//
// The function takes an sales ID and a warehouse ID as input and updates the stock status of the sales
// order to "updated". The function also adds stock movements for the items in the sales order.
//
// The function returns an error if the operation fails.
func (s *SalesService) UpdateStock(salesID, warehouseID string, description string) error {
	var sales models.SalesModel
	if err := s.db.First(&sales, salesID).Error; err != nil {
		return err
	}

	invSrv, ok := s.ctx.InventoryService.(*inventory.InventoryService)
	if !ok {
		return errors.New("invalid inventory service")
	}

	// Pastikan status PO adalah "pending"
	if sales.StockStatus != "pending" {
		return errors.New("purchase order already processed")
	}
	return s.ctx.DB.Transaction(func(tx *gorm.DB) error {
		for _, v := range sales.Items {
			if v.ProductID == nil || v.WarehouseID == nil {
				continue
			}
			_, err := invSrv.StockMovementService.AddMovement(sales.SalesDate, *v.ProductID, *v.WarehouseID, v.VariantID, nil, nil, nil, -v.Quantity, models.MovementTypeIn, sales.ID, description)
			if err != nil {
				tx.Rollback()
				return err
			}
		}

		sales.StockStatus = "updated"

		if err := tx.Save(&sales).Error; err != nil {
			tx.Rollback()
			return err
		}

		// return nil will commit the whole transaction
		return nil
	})

}

// CreateSalesFromOrderRequest creates a new sales document from an order request.
//
// The function takes an order request, a sales number, a tax percentage, and a description as input.
// It then creates a new sales document based on the order request and its items. The function
// returns an error if the operation fails.
//
// The function uses the tax percentage to calculate the total before tax and the total after
// tax. The function also uses the discount percent and discount amount to calculate the total
// before discount and the total after discount.
//
// The function returns an error if the operation fails.
func (s *SalesService) CreateSalesFromOrderRequest(orderRequest *models.OrderRequestModel, salesNumber string, taxPercent float64, description string) error {
	var companyID *string
	if s.ctx.Request.Header.Get("ID-Company") != "" {
		compID := s.ctx.Request.Header.Get("ID-Company")
		companyID = &compID
	}
	if orderRequest.ContactID == nil {
		return errors.New("contact ID is required")
	}
	contactData, err := json.Marshal(*orderRequest.Contact)
	if err != nil {
		return err
	}

	data := &models.SalesModel{
		SalesNumber:     salesNumber,
		Code:            utils.RandString(10, true),
		SalesDate:       *orderRequest.CreatedAt,
		DueDate:         &orderRequest.ExpiresAt,
		TotalBeforeTax:  0,
		TotalBeforeDisc: 0,
		Subtotal:        0,
		Paid:            0,
		CompanyID:       companyID,
		ContactID:       orderRequest.ContactID,
		ContactData:     string(contactData),
		Type:            models.ONLINE,
		Items:           []models.SalesItemModel{},
	}
	var totalBeforeTax, totalBeforeDisc float64
	for _, v := range orderRequest.Items {
		data.Items = append(data.Items, models.SalesItemModel{
			ProductID:          v.ProductID,
			Quantity:           v.Quantity,
			UnitPrice:          v.UnitPrice,
			Total:              v.Quantity * v.UnitPrice,
			DiscountPercent:    v.DiscountPercent,
			DiscountAmount:     v.DiscountAmount,
			SubtotalBeforeDisc: v.Quantity * v.UnitPrice,
		})
		totalBeforeDisc += v.Quantity * v.UnitPrice
		if v.DiscountPercent > 0 {
			totalBeforeTax += v.Quantity * (v.UnitPrice - (v.UnitPrice * v.DiscountPercent / 100))
		} else {
			totalBeforeTax += v.Quantity * (v.UnitPrice - v.DiscountAmount)
		}

	}
	data.TotalBeforeTax = totalBeforeTax
	data.TotalBeforeDisc = totalBeforeDisc
	data.Subtotal = totalBeforeTax * (1 + taxPercent/100)
	return s.CreateSales(data)
}

// AddItem adds a new item to a sales document.
//
// It takes a sales document and a new item as input, and returns an error if the operation fails.
//
// The function first loads the product associated with the item, and sets the item's base price
// to the product's price. If the product has a tax set, the item is also set to have the same tax.
//
// The function then creates the item in the database, and returns an error if the operation fails.
//
// Finally, the function calls the UpdateTotal function to recalculate the total of the sales document.
func (s *SalesService) AddItem(sales *models.SalesModel, item *models.SalesItemModel) error {
	if item.ProductID != nil {
		var product models.ProductModel
		if err := s.db.Select("id, price").First(&product, "id = ?", *item.ProductID).Error; err != nil {
			return err
		}
		item.BasePrice = product.Price

		if product.TaxID != nil {
			item.TaxID = product.TaxID
			item.Tax = product.Tax
		}
	}

	err := s.db.Create(item).Error
	if err != nil {
		return err
	}

	return s.UpdateTotal(sales)
}

// GetItems returns all items associated with a sales document.
//
// It takes a sales ID as input, and returns a slice of items and an error if the operation fails.
//
// The function uses the GORM preload function to load all items associated with the sales document,
// as well as the associated product, unit, variant, warehouse, sale account, asset account, and tax.
func (s *SalesService) GetItems(id string) ([]models.SalesItemModel, error) {
	var items []models.SalesItemModel

	err := s.db.
		Preload("Product", func(db *gorm.DB) *gorm.DB {
			return db.Preload("Category")
		}).
		Preload("Unit").
		Preload("Variant").
		Preload("Warehouse").
		Preload("SaleAccount").
		Preload("AssetAccount").
		Preload("Tax").
		Where("sales_id = ?", id).Order("created_at ASC").Find(&items).Error
	if err != nil {
		return nil, err
	}

	for i, v := range items {
		if v.ProductID != nil {
			v.Product.GetPrices(s.db)
		}
		items[i] = v
	}

	return items, nil
}

// UpdateTotal recalculates the total of a sales document.
//
// It takes a sales document as input, and returns an error if the operation fails.
//
// The function first loads all items associated with the sales document, and then calculates the total
// before tax, the total before discount, the total tax, and the total discount.
//
// The function then calculates the total after tax by calling the CalculateTaxes function, and
// sets the total of the sales document to the calculated total.
//
// Finally, the function updates the sales document in the database, and returns an error if the operation fails.
func (s *SalesService) UpdateTotal(sales *models.SalesModel) error {
	s.db.Preload("Items").Model(sales).Find(sales)
	var totalBeforeTax, totalBeforeDisc, subTotal, itemsTax, totalDisc float64
	for _, v := range sales.Items {
		totalBeforeDisc += v.SubtotalBeforeDisc
		totalBeforeTax += v.SubTotal
		subTotal += v.SubTotal
		itemsTax += v.TotalTax
		totalDisc += v.DiscountAmount
	}
	sales.TotalBeforeTax = totalBeforeTax
	sales.TotalBeforeDisc = totalBeforeDisc
	sales.Subtotal = subTotal

	afterTax, salesTaxAmount, taxBreakdown := s.CalculateTaxes(subTotal, sales.IsCompound, sales.Taxes)
	sales.Subtotal = afterTax
	sales.TotalTax = itemsTax + salesTaxAmount
	sales.Total = sales.Subtotal + sales.TotalTax
	sales.TotalDiscount = totalDisc
	b, _ := json.Marshal(taxBreakdown)
	sales.TaxBreakdown = string(b)

	return s.db.Omit(clause.Associations).Save(&sales).Error
}

// DeleteItem deletes an item from a sales document.
//
// It takes a sales document and an item ID as input, and returns an error if the operation fails.
//
// The function first deletes the item from the database, and then calls the UpdateTotal function
// to recalculate the total of the sales document.
func (s *SalesService) DeleteItem(sales *models.SalesModel, itemID string) error {
	err := s.db.Where("sales_id = ? AND id = ?", sales.ID, itemID).Delete(&models.SalesItemModel{}).Error
	if err != nil {
		return err
	}
	return s.UpdateTotal(sales)
}

// UpdateItem updates an item in a sales document.
//
// It takes a sales document, an item ID, and an updated item as input, and returns an error if the operation fails.
//
// The function first loads the product associated with the item, and sets the item's base price
// to the product's price. If the product has a tax set, the item is also set to have the same tax.
//
// The function then updates the item in the database, and returns an error if the operation fails.
//
// Finally, the function calls the UpdateTotal function to recalculate the total of the sales document.
func (s *SalesService) UpdateItem(sales *models.SalesModel, itemID string, item *models.SalesItemModel) error {
	if item.ProductID != nil {
		var product models.ProductModel
		if err := s.db.Select("id, price").First(&product, "id = ?", *item.ProductID).Error; err != nil {
			return err
		}
		item.BasePrice = product.Price
	}
	taxPercent := 0.0
	taxAmount := 0.0

	if item.UnitID != nil {
		productUnit := models.ProductUnitData{}
		s.db.Model(&productUnit).Where("product_model_id = ? and unit_model_id = ?", item.ProductID, item.UnitID).Find(&productUnit)
		item.UnitValue = productUnit.Value
	} else {
		item.UnitValue = 1
	}

	if item.TaxID != nil {
		taxPercent = item.Tax.Amount
	}
	item.SubtotalBeforeDisc = (item.Quantity * item.UnitValue) * item.UnitPrice
	if item.DiscountPercent > 0 {
		taxAmount = (item.SubtotalBeforeDisc - (item.SubtotalBeforeDisc * item.DiscountPercent / 100)) * (taxPercent / 100)
		item.SubTotal = (item.SubtotalBeforeDisc - (item.SubtotalBeforeDisc * item.DiscountPercent / 100))
		item.DiscountAmount = item.SubtotalBeforeDisc * item.DiscountPercent / 100
	} else {
		taxAmount = (item.SubtotalBeforeDisc - item.DiscountAmount) * (taxPercent / 100)
		item.SubTotal = (item.SubtotalBeforeDisc - item.DiscountAmount)
		item.DiscountPercent = 0
	}
	item.TotalTax = taxAmount
	item.Total = item.SubTotal + taxAmount
	err := s.db.Where("sales_id = ? AND id = ?", sales.ID, itemID).Omit("sales_id").Save(item).Error
	if err != nil {
		return err
	}
	return s.UpdateTotal(sales)
}

// CalculateTaxes computes the total amount after applying taxes to a base amount.
//
// This function iterates over a list of tax models, calculates the tax for each model,
// and aggregates the total tax amount. If the isCompound flag is true, it applies each
// tax incrementally, recalculating the base amount after each tax application. If false,
// it applies the total tax at once to the base amount. The function returns the total
// amount after taxes, the total tax amount, and a breakdown of individual tax amounts.
//
// Parameters:
//   - baseAmount: The initial amount before taxes.
//   - isCompound: A boolean indicating whether taxes should be compounded.
//   - taxes: A slice of TaxModel pointers representing the taxes to be applied.
//
// Returns:
//   - The total amount after all taxes have been applied.
//   - The aggregated total tax amount.
//   - A map detailing the tax amount for each tax model by name.

func (s *SalesService) CalculateTaxes(baseAmount float64, isCompound bool, taxes []*models.TaxModel) (float64, float64, map[string]float64) {
	totalAmount := baseAmount
	taxBreakdown := make(map[string]float64)
	totalTax := 0.0
	for _, tax := range taxes {
		if tax == nil {
			continue
		}
		taxAmount := (totalAmount * tax.Amount) / 100
		// fmt.Printf("TAX_AMOUNT CalculateTaxes %f\n", taxAmount)
		totalTax += taxAmount
		taxBreakdown[tax.Name] = taxAmount

		if isCompound {
			totalAmount += taxAmount
		}
	}

	if !isCompound {
		totalAmount += totalTax
	}
	return totalAmount, totalTax, taxBreakdown
}

// PublishSales creates a new sales document in the database and performs relevant accounting entries.
//
// If the sales document has items with a sale account and/or an asset account, transactions will be created
// for the sale and the asset account. If the sales document has a payment account, the sales document will be
// marked as paid. If the sales document has a partial payment, the sales document will be marked as partial.
//
// If successful, it returns nil; otherwise, it returns an error indicating what went wrong.
func (s *SalesService) PublishSales(data *models.SalesModel) error {
	if len(data.Items) == 0 {
		return errors.New("sales has no items")
	}
	now := time.Now()
	data.PublishedAt = &now
	if s.financeService.TransactionService == nil {
		return errors.New("transaction service is not set")
	}
	if err := s.db.Save(data).Error; err != nil {
		return err
	}
	if data.DocumentType != "INVOICE" {
		return nil
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		for _, v := range data.Items {
			if v.SaleAccountID != nil {
				s.financeService.TransactionService.CreateTransaction(&models.TransactionModel{
					Date:               data.SalesDate,
					AccountID:          v.SaleAccountID,
					Description:        "Penjualan " + data.SalesNumber,
					Notes:              data.Description,
					TransactionRefID:   &data.ID,
					TransactionRefType: "sales",
					CompanyID:          data.CompanyID,
					Credit:             v.SubTotal,
				}, v.Total)
			}
			if v.AssetAccountID != nil {
				s.financeService.TransactionService.CreateTransaction(&models.TransactionModel{
					Date:               data.SalesDate,
					AccountID:          v.AssetAccountID,
					Description:        "Penjualan " + data.SalesNumber,
					Notes:              data.Description,
					TransactionRefID:   &data.ID,
					TransactionRefType: "sales",
					CompanyID:          data.CompanyID,
					Debit:              v.SubTotal,
				}, v.Total)
			}

			if v.TaxID != nil {
				// HUTANG PAJAK
				s.financeService.TransactionService.CreateTransaction(&models.TransactionModel{
					Date:               data.SalesDate,
					AccountID:          v.SaleAccountID,
					Description:        "Pajak Penjualan " + data.SalesNumber,
					Notes:              data.Description,
					TransactionRefID:   &data.ID,
					TransactionRefType: "sales",
					CompanyID:          data.CompanyID,
					Credit:             v.TotalTax,
				}, v.Total)

				// ASET / PIUTANG PAJAK
				s.financeService.TransactionService.CreateTransaction(&models.TransactionModel{
					Date:               data.SalesDate,
					AccountID:          v.AssetAccountID,
					Description:        "Pajak Penjualan " + data.SalesNumber,
					Notes:              data.Description,
					TransactionRefID:   &data.ID,
					TransactionRefType: "sales",
					CompanyID:          data.CompanyID,
					Debit:              v.TotalTax,
				}, v.Total)
			}

			if v.ProductID != nil {
				if v.WarehouseID == nil {
					return errors.New("warehouse ID is required")
				}
				// ADD MOVEMENT
				_, err := s.inventoryService.StockMovementService.AddMovement(
					time.Now(),
					*v.ProductID,
					*v.WarehouseID,
					v.VariantID,
					nil,
					nil,
					nil,
					-v.Quantity,
					models.MovementTypeSale,
					data.ID,
					fmt.Sprintf("Sales #%s", data.SalesNumber))
				if err != nil {
					return err
				}
				// ADD SUPPLY TRANSACTION
			}
		}
		return nil
	})
}

// PostInvoice posts a sales invoice with the given ID and data, and updates the status of the invoice to "POSTED".
//
// The function takes a pointer to a SalesModel and a string representing the user ID.
// It verifies that the document type is "INVOICE" and that there are items present in the sales model.
// It updates the status of the invoice to "POSTED", sets the published at and published by fields, and manages payment terms if applicable.
// It retrieves the necessary accounts for cost of goods sold (COGS) and inventory, and creates financial transactions for each item in the sales model.
// It also manages stock movements for products associated with the invoice.
// The function executes these operations within a transaction to ensure data consistency.
// Returns an error if any of the operations fail.
func (s *SalesService) PostInvoice(id string, data *models.SalesModel, userID string, date time.Time) error {

	if data.DocumentType != "INVOICE" {
		return errors.New("document type is not invoice")
	}

	if len(data.Items) == 0 {
		return errors.New("items is required")
	}
	now := time.Now()

	if data.PaymentTermsCode != "" {
		var paymentTerms models.PaymentTermModel

		err := s.db.Find(&paymentTerms, "code = ?", data.PaymentTermsCode).Error
		if err == nil && data.DueDate == nil && paymentTerms.DueDays != nil {
			due := date.AddDate(0, 0, *paymentTerms.DueDays)
			data.DueDate = &due
		}
		if err == nil && paymentTerms.DiscountDueDays != nil {
			due := date.AddDate(0, 0, *paymentTerms.DiscountDueDays)
			data.DiscountDueDate = &due
			data.PaymentDiscountAmount = *paymentTerms.DiscountAmount
		}
	}

	data.Status = "POSTED"
	data.PublishedAt = &now
	data.PublishedByID = &userID
	refType := "sales"
	secRefType := "sales_item"

	// GET COGS ACCOUNT
	var cogsAccount models.AccountModel
	err := s.db.Where("is_cogs_account = ? and company_id = ?", true, *data.CompanyID).First(&cogsAccount).Error
	if err != nil {
		return errors.New("cogs account not found")
	}
	// GET INVENTORY ACCOUNT
	var inventoryAccount models.AccountModel
	err = s.db.Where("is_inventory_account = ? and company_id = ?", true, *data.CompanyID).First(&inventoryAccount).Error
	if err != nil {
		return errors.New("inventory account not found")
	}
	if data.PaymentAccount.Type == "ASSET" {
		data.Paid = data.Total

	}
	assetID := utils.Uuid()
	err = s.db.Transaction(func(tx *gorm.DB) error {
		s.financeService.TransactionService.SetDB(tx)
		s.inventoryService.StockMovementService.SetDB(tx)
		totalPayment := 0.0
		for _, v := range data.Items {
			if v.SaleAccountID == nil {
				return errors.New("sale account ID is required")
			}
			err := s.financeService.TransactionService.CreateTransaction(&models.TransactionModel{
				Date:                        date,
				AccountID:                   v.SaleAccountID,
				Description:                 "Penjualan " + data.SalesNumber,
				Notes:                       v.Description,
				TransactionRefID:            &assetID,
				TransactionRefType:          "transaction",
				TransactionSecondaryRefID:   &data.ID,
				TransactionSecondaryRefType: refType,
				CompanyID:                   data.CompanyID,
				Credit:                      v.SubTotal,
				UserID:                      &userID,
				IsIncome:                    true,
			}, v.SubTotal)
			if err != nil {
				return err
			}
			// if v.AssetAccountID != nil {
			// 	err = s.financeService.TransactionService.CreateTransaction(&models.TransactionModel{
			// 		Date:                        date,
			// 		AccountID:                   v.AssetAccountID,
			// 		Description:                 "Penjualan " + data.SalesNumber,
			// 		Notes:                       v.Description,
			// 		TransactionRefID:            &data.ID,
			// 		TransactionRefType:          refType,
			// 		TransactionSecondaryRefID:   &v.ID,
			// 		TransactionSecondaryRefType: secRefType,
			// 		CompanyID:                   data.CompanyID,
			// 		Debit:                       v.SubTotal + v.TotalTax,
			// 		UserID:                      &userID,
			// 	}, v.SubTotal+v.TotalTax)
			// 	if err != nil {
			// 		return err
			// 	}
			// } else {
			totalPayment += v.SubTotal + v.TotalTax
			// }

			if v.TaxID != nil {
				if v.Tax == nil {
					return errors.New("tax is required")
				}
				// HUTANG PAJAK
				err := s.financeService.TransactionService.CreateTransaction(&models.TransactionModel{
					Date:                        date,
					AccountID:                   v.Tax.AccountPayableID,
					Description:                 "Hutang Pajak " + data.SalesNumber,
					Notes:                       v.Description,
					TransactionRefID:            &data.ID,
					TransactionRefType:          refType,
					TransactionSecondaryRefID:   &v.ID,
					TransactionSecondaryRefType: secRefType,
					CompanyID:                   data.CompanyID,
					Credit:                      v.TotalTax,
					UserID:                      &userID,
					IsAccountPayable:            true,
					IsTax:                       true,
				}, v.TotalTax)
				if err != nil {
					return err
				}

			}

			if v.ProductID != nil {
				if v.WarehouseID == nil {
					return errors.New("warehouse ID is required")
				}
				// ADD MOVEMENT
				movement, err := s.inventoryService.StockMovementService.AddMovement(
					time.Now(),
					*v.ProductID,
					*v.WarehouseID,
					v.VariantID,
					nil,
					nil,
					data.CompanyID,
					-v.Quantity,
					models.MovementTypeSale,
					data.ID,
					fmt.Sprintf("Sales %s (%s)", data.SalesNumber, v.Description))
				if err != nil {
					return err
				}
				movement.ReferenceID = data.ID
				movement.ReferenceType = &refType
				movement.SecondaryRefID = &v.ID
				movement.SecondaryRefType = &secRefType
				movement.Value = v.UnitValue
				movement.UnitID = v.UnitID

				err = tx.Save(movement).Error
				if err != nil {
					return err
				}
				// ADD SUPPLY TRANSACTION
				err = s.financeService.TransactionService.CreateTransaction(&models.TransactionModel{
					Date:                        date,
					AccountID:                   &inventoryAccount.ID,
					Description:                 "Persediaan " + data.SalesNumber,
					Notes:                       v.Description,
					TransactionRefID:            &movement.ID,
					TransactionRefType:          "stock_movement",
					TransactionSecondaryRefID:   &data.ID,
					TransactionSecondaryRefType: refType,
					CompanyID:                   data.CompanyID,
					Credit:                      v.BasePrice * v.Quantity * v.UnitValue,
					UserID:                      &userID,
				}, v.BasePrice*v.Quantity*v.UnitValue)
				if err != nil {
					return err
				}

				// ADD COGS TRANSACTION
				err = s.financeService.TransactionService.CreateTransaction(&models.TransactionModel{
					Date:                        date,
					AccountID:                   &cogsAccount.ID,
					Description:                 "HPP " + data.SalesNumber,
					Notes:                       v.Description,
					TransactionRefID:            &movement.ID,
					TransactionRefType:          "stock_movement",
					TransactionSecondaryRefID:   &data.ID,
					TransactionSecondaryRefType: refType,
					CompanyID:                   data.CompanyID,
					Debit:                       v.BasePrice * v.Quantity * v.UnitValue,
					UserID:                      &userID,
				}, v.BasePrice*v.Quantity*v.UnitValue)
				if err != nil {
					return err
				}

			}

			err = tx.Save(v).Error
			if err != nil {
				return err
			}
		}

		err = s.financeService.TransactionService.CreateTransaction(&models.TransactionModel{
			BaseModel:          shared.BaseModel{ID: assetID},
			Date:               date,
			AccountID:          data.PaymentAccountID,
			Description:        "Penjualan " + data.SalesNumber,
			TransactionRefID:   &data.ID,
			TransactionRefType: refType,
			CompanyID:          data.CompanyID,
			Debit:              totalPayment,
			UserID:             &userID,
		}, totalPayment)
		return tx.Save(data).Error
	})
	s.financeService.TransactionService.SetDB(s.db)
	s.inventoryService.StockMovementService.SetDB(s.db)
	return err
}

// GetBalance calculates the remaining balance of a sales order.
//
// If the payment account is an asset account, it returns 0 immediately.
// Otherwise, it calculates the total payment amount made to the sales order,
// and returns the difference between the sales order total and the total payment.
// If the payment is more than the total, it returns an error.
func (s *SalesService) GetBalance(sales *models.SalesModel) (float64, error) {
	if sales.PaymentAccount.Type == "ASSET" {
		return 0, nil
	}
	var amount struct {
		Sum float64 `sql:"sum"`
	}
	err := s.db.Model(&models.SalesPaymentModel{}).Where("sales_id = ?", sales.ID).Select("sum(amount)").Scan(&amount).Error
	if err != nil {
		return 0, err
	}
	if sales.Total > amount.Sum {
		return sales.Total - amount.Sum, nil
	}
	return 0, errors.New("payment is more than total")
}

// CreateSalesPayment creates a new sales payment transaction associated with a sales order.
//
// It performs the following operations:
//  1. Verifies if the payment amount is greater than the remaining balance of the sales order. If so, it returns an error.
//  2. Retrieves the payment account associated with the sales order and verifies its type. If the account type is not RECEIVABLE, it returns an error.
//  3. Creates a new transaction record in the database with the following details:
//     - Date: the provided date
//     - AccountID: the ID of the receivable account associated with the sales order
//     - Description: "Pembayaran [sales number]"
//     - Notes: the description of the sales order
//     - TransactionRefID: the ID of the asset account associated with the payment
//     - TransactionRefType: "transaction"
//     - CompanyID: the ID of the company associated with the sales order
//     - Credit: the payment amount
//  4. Creates another transaction record with the following details:
//     - AccountID: the ID of the asset account associated with the payment
//     - Debit: the payment amount
//  5. If the payment discount is greater than 0, it creates another transaction record with the following details:
//     - AccountID: the ID of the contra revenue account associated with the company
//     - Debit: the discount amount
//  6. Saves the sales payment data in the database.
//  7. Commits the transaction if all operations are successful. Otherwise, it rolls back the transaction.
//
// Returns an error if any of the operations fail.
func (s *SalesService) CreateSalesPayment(sales *models.SalesModel, salesPayment *models.SalesPaymentModel) error {
	err := s.db.Transaction(func(tx *gorm.DB) error {
		s.financeService.TransactionService.SetDB(tx)
		balance, err := s.GetBalance(sales)
		if err != nil {
			return err
		}
		if balance < salesPayment.Amount {
			return errors.New("payment is more than balance")
		}

		if salesPayment.AssetAccountID == nil {
			return errors.New("asset account is required")
		}
		if sales.PaymentAccountID == nil {
			return errors.New("sales payment account not found")
		}

		if sales.PaymentAccount.Type != "RECEIVABLE" {
			return errors.New("sales payment account type must be RECEIVABLE")
		}
		paymentAmount := salesPayment.Amount
		discountAmount := 0.0
		if salesPayment.PaymentDiscount > 0 {
			paymentAmount = salesPayment.Amount - (salesPayment.Amount * (salesPayment.PaymentDiscount / 100))
			discountAmount = salesPayment.Amount * (salesPayment.PaymentDiscount / 100)
		}

		paymentID := uuid.New().String()
		receivableID := uuid.New().String()
		assetTransID := uuid.New().String()

		receivableData := models.TransactionModel{
			BaseModel:                   shared.BaseModel{ID: receivableID},
			Date:                        salesPayment.PaymentDate,
			AccountID:                   sales.PaymentAccountID,
			Description:                 "Pembayaran " + sales.SalesNumber,
			Notes:                       salesPayment.Notes,
			TransactionRefID:            &assetTransID,
			TransactionRefType:          "transaction",
			CompanyID:                   sales.CompanyID,
			Credit:                      salesPayment.Amount,
			UserID:                      salesPayment.UserID,
			TransactionSecondaryRefID:   &sales.ID,
			TransactionSecondaryRefType: "sales",
		}
		receivableData.ID = receivableID
		err = s.financeService.TransactionService.CreateTransaction(&receivableData, salesPayment.Amount)
		if err != nil {
			return err
		}

		assetData := models.TransactionModel{
			BaseModel:                   shared.BaseModel{ID: assetTransID},
			Date:                        salesPayment.PaymentDate,
			AccountID:                   salesPayment.AssetAccountID,
			Description:                 "Pembayaran " + sales.SalesNumber,
			Notes:                       salesPayment.Notes,
			TransactionRefID:            &receivableData.ID,
			TransactionRefType:          "transaction",
			CompanyID:                   sales.CompanyID,
			Debit:                       paymentAmount,
			UserID:                      salesPayment.UserID,
			TransactionSecondaryRefID:   &sales.ID,
			TransactionSecondaryRefType: "sales",
		}

		assetData.ID = assetTransID
		err = s.financeService.TransactionService.CreateTransaction(&assetData, paymentAmount)
		if err != nil {
			return err
		}

		if discountAmount > 0 {
			var contraRevenueAccount models.AccountModel
			err := s.db.Where("type = ? and company_id = ? and is_discount = ?", models.CONTRA_REVENUE, sales.CompanyID, true).First(&contraRevenueAccount).Error
			if err != nil {
				return err
			}
			err = s.financeService.TransactionService.CreateTransaction(&models.TransactionModel{
				Date:                        salesPayment.PaymentDate,
				AccountID:                   &contraRevenueAccount.ID,
				Description:                 "Diskon " + sales.SalesNumber,
				TransactionRefID:            &receivableData.ID,
				TransactionRefType:          "transaction",
				CompanyID:                   sales.CompanyID,
				Debit:                       discountAmount,
				UserID:                      salesPayment.UserID,
				TransactionSecondaryRefID:   &sales.ID,
				TransactionSecondaryRefType: "sales",
			}, discountAmount)
			if err != nil {
				return err
			}
		}

		salesPayment.ID = paymentID

		return tx.Create(salesPayment).Error
	})
	s.financeService.TransactionService.SetDB(s.db)
	return err
}

// GetPdf generates a PDF invoice for a sales order.
//
// It accepts the sales order, a template path, a time format string, a footer string,
// and flags indicating whether to show the company and shipping information.
// The function retrieves the company information from the database and formats the invoice
// with the sales order details, items, payments, and contact information.
// It returns a byte slice of the generated PDF or an error if any operation fails.
func (s *SalesService) GetPdf(sales *models.SalesModel, templatePath, timeFormatStr, footer string, showCompany, showShipped bool) ([]byte, error) {
	if timeFormatStr == "" {
		timeFormatStr = "02/01/2006"
	}

	company := models.CompanyModel{}
	err := s.db.Find(&company, "id = ?", *sales.CompanyID).Error
	if err != nil {
		return nil, err
	}

	dueDate := ""
	if sales.DueDate != nil {
		dueDate = sales.DueDate.Format(timeFormatStr)
	}

	items := []utils.InvoicePDFItem{}
	for i, v := range sales.Items {
		unitName := ""
		if v.Unit != nil {
			unitName = v.Unit.Name
		}
		taxPercent := 0.0
		taxName := ""
		if v.Tax != nil {
			taxPercent = v.Tax.Amount
			taxName = v.Tax.Code
		}
		items = append(items, utils.InvoicePDFItem{
			No:                 i + 1,
			Description:        v.Description,
			Notes:              v.Notes,
			Quantity:           utils.FormatRupiah(v.Quantity),
			UnitPrice:          utils.FormatRupiah(v.UnitPrice),
			UnitName:           unitName,
			Total:              utils.FormatRupiah(v.Total),
			SubTotal:           utils.FormatRupiah(v.SubTotal),
			SubtotalBeforeDisc: utils.FormatRupiah(v.SubtotalBeforeDisc),
			TotalDiscount:      utils.FormatRupiah(v.DiscountAmount),
			DiscountPercent:    utils.FormatRupiah(v.DiscountPercent),
			TaxAmount:          utils.FormatRupiah(v.TotalTax),
			TaxPercent:         utils.FormatRupiah(taxPercent),
			TaxName:            taxName,
		})
	}

	payments := []utils.InvoicePDFPayment{}

	for _, v := range sales.SalesPayments {
		payments = append(payments, utils.InvoicePDFPayment{
			Date:               v.PaymentDate.Format(timeFormatStr),
			Description:        v.Notes,
			PaymentMethod:      strings.ReplaceAll(v.PaymentMethod, "_", " "),
			Amount:             utils.FormatRupiah(v.Amount),
			PaymentDiscount:    utils.FormatRupiah(v.PaymentDiscount),
			PaymentMethodNotes: v.PaymentMethodNotes,
		})
	}

	billedTo := utils.InvoicePDFContact{}
	shippedTo := utils.InvoicePDFContact{}
	contactName, ok := sales.ContactDataParsed["name"].(string)
	if ok {
		billedTo.Name = contactName
	}
	contactEmail, ok := sales.ContactDataParsed["email"].(string)
	if ok {
		billedTo.Email = contactEmail
	}
	contactPhone, ok := sales.ContactDataParsed["phone"].(string)
	if ok {
		billedTo.Phone = contactPhone
	}
	contactAddress, ok := sales.ContactDataParsed["address"].(string)
	if ok {
		billedTo.Address = contactAddress
	}

	shippedName, ok := sales.DeliveryDataParsed["name"].(string)
	if ok {
		shippedTo.Name = shippedName
	}
	shippedEmail, ok := sales.DeliveryDataParsed["email"].(string)
	if ok {
		shippedTo.Email = shippedEmail
	}
	shippedPhone, ok := sales.DeliveryDataParsed["phone"].(string)
	if ok {
		shippedTo.Phone = shippedPhone
	}
	shippedAddress, ok := sales.DeliveryDataParsed["address"].(string)
	if ok {
		shippedTo.Address = shippedAddress
	}

	var data = utils.InvoicePDF{
		ShowCompany: showCompany,
		ShowShipped: showShipped,
		Company: utils.InvoicePDFContact{
			Name:    company.Name,
			Address: company.Address,
			Phone:   company.Phone,
			Email:   company.Email,
		},
		Number:          sales.SalesNumber,
		Date:            sales.SalesDate.Format(timeFormatStr),
		DueDate:         dueDate,
		Items:           items,
		SubTotal:        utils.FormatRupiah(sales.Subtotal),
		TotalDiscount:   utils.FormatRupiah(sales.TotalDiscount),
		AfterDiscount:   utils.FormatRupiah(sales.Total - sales.TotalDiscount),
		TotalTax:        utils.FormatRupiah(sales.TotalTax),
		GrandTotal:      utils.FormatRupiah(sales.Total),
		InvoicePayments: payments,
		Balance:         utils.FormatRupiah(sales.Total - sales.Paid),
		Paid:            utils.FormatRupiah(sales.Paid),
		BilledTo:        billedTo,
		ShippedTo:       shippedTo,
		TermCondition:   sales.TermCondition,
		PaymentTerms:    sales.PaymentTerms,
	}

	return utils.GenerateInvoicePDF(data, templatePath, footer)
}
