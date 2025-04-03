package sales

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/finance"
	"github.com/AMETORY/ametory-erp-modules/inventory"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
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

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&models.SalesModel{}, &models.SalesItemModel{})
}

func NewSalesService(db *gorm.DB, ctx *context.ERPContext, financeService *finance.FinanceService, inventoryService *inventory.InventoryService) *SalesService {
	return &SalesService{db: db, ctx: ctx, financeService: financeService, inventoryService: inventoryService}
}

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

		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			return err
		}

		return nil
	})

}

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

		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			return err
		}

		return nil
	})
}

func (s *SalesService) UpdateSales(id string, data *models.SalesModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

func (s *SalesService) DeleteSales(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.SalesModel{}).Error
}

func (s *SalesService) GetSalesByID(id string) (*models.SalesModel, error) {
	var sales models.SalesModel
	err := s.db.Where("id = ?", id).First(&sales).Error
	return &sales, err
}

func (s *SalesService) GetSalesByCode(code string) (*models.SalesModel, error) {
	var sales models.SalesModel
	err := s.db.Where("code = ?", code).First(&sales).Error
	return &sales, err
}

func (s *SalesService) GetSalesBySalesNumber(salesNumber string) (*models.SalesModel, error) {
	var sales models.SalesModel
	err := s.db.Where("sales_number = ?", salesNumber).First(&sales).Error
	return &sales, err
}

func (s *SalesService) GetSales(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.Preload("Contact")
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
	stmt = stmt.Model(&models.SalesModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.SalesModel{})
	page.Page = page.Page + 1
	return page, nil
}

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
	err := s.ctx.DB.Transaction(func(tx *gorm.DB) error {
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

		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			return err
		}

		// return nil will commit the whole transaction
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

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
		SalesDate:       orderRequest.CreatedAt,
		DueDate:         orderRequest.ExpiresAt,
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

func (s *SalesService) AddItem(sales *models.SalesModel, item *models.SalesItemModel) error {

	err := s.db.Create(item).Error
	if err != nil {
		return err
	}
	return s.UpdateTotal(sales)
}

func (s *SalesService) DeleteItem(sales *models.SalesModel, itemID string) error {
	err := s.db.Where("sales_id = ? AND id = ?", sales.ID, itemID).Delete(&models.SalesItemModel{}).Error
	if err != nil {
		return err
	}
	return s.UpdateTotal(sales)
}

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
	// fmt.Printf("TAX_AMOUNT %f", salesTaxAmount)
	sales.Subtotal = afterTax
	sales.TotalTax = itemsTax + salesTaxAmount
	sales.Total = sales.Subtotal + sales.TotalTax
	sales.TotalDiscount = totalDisc
	b, _ := json.Marshal(taxBreakdown)
	sales.TaxBreakdown = string(b)

	return s.db.Omit(clause.Associations).Save(&sales).Error
}

func (s *SalesService) UpdateItem(sales *models.SalesModel, itemID string, item *models.SalesItemModel) error {
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
