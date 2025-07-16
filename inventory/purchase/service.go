package purchase

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/finance"
	stockmovement "github.com/AMETORY/ametory-erp-modules/inventory/stock_movement"
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/google/uuid"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PurchaseService struct {
	db                   *gorm.DB
	ctx                  *context.ERPContext
	financeService       *finance.FinanceService
	stockMovementService *stockmovement.StockMovementService
}

// NewPurchaseService creates a new instance of PurchaseService with the given database connection, context, finance service and stock movement service.
func NewPurchaseService(db *gorm.DB, ctx *context.ERPContext, financeService *finance.FinanceService, stockMovementService *stockmovement.StockMovementService) *PurchaseService {
	return &PurchaseService{
		db:                   db,
		ctx:                  ctx,
		financeService:       financeService,
		stockMovementService: stockMovementService,
	}
}

// Migrate migrates the purchase database model to the given database connection.
//
// It uses gorm's AutoMigrate method to create the tables if they don't exist, and to migrate the existing tables if they do.
//
// AutoMigrate will add missing columns, but won't change existing column's type or delete unused column, it also won't delete/rename tables.
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&models.PurchaseOrderModel{}, &models.PurchaseOrderItemModel{}, &models.PurchasePaymentModel{})
}

// UpdatePurchase updates the purchase order with the given id with the given data.
//
// It takes the id of the purchase order to be updated and a pointer to a PurchaseOrderModel which contains the updated data of the purchase order.
// The function returns an error if the update operation fails.
func (s *PurchaseService) UpdatePurchase(id string, data *models.PurchaseOrderModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

func (s *PurchaseService) DeletePurchase(id string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		fmt.Println("DELETE PURCHASE")
		err := tx.Where("transaction_ref_id = ? OR transaction_secondary_ref_id = ?", id, id).Delete(&models.TransactionModel{}).Error
		if err != nil {
			return err
		}
		err = tx.Where("purchase_id = ?", id).Delete(&models.PurchaseOrderItemModel{}).Error
		if err != nil {
			return err
		}
		err = tx.Where("reference_id = ? or secondary_ref_id = ?", id, id).Delete(&models.StockMovementModel{}).Error
		if err != nil {
			return err
		}
		return tx.Where("id = ?", id).Delete(&models.PurchaseOrderModel{}).Error
	})

}

// ClearTransaction clears all transaction with given id as reference id.
//
// This function runs inside a transaction, so if any error occurs, it will rollback.
//
// It uses gorm's Delete method to delete all transaction with given id as reference id.
func (s *PurchaseService) ClearTransaction(id string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		return tx.Where("transaction_ref_id = ?", id).Delete(&models.TransactionModel{}).Error
	})

}

// CreatePurchaseOrder creates a new purchase order in the database.
//
// It accepts a pointer to a PurchaseOrderModel which contains the purchase order details.
// The function retrieves the company ID from the request header and verifies the existence
// of an inventory account associated with the company. If the inventory account is not found,
// it returns an error. Otherwise, it saves the purchase order data in the database.
//
// Returns an error if the creation of the purchase order fails or if the inventory account
// is not found.

func (s *PurchaseService) CreatePurchaseOrder(data *models.PurchaseOrderModel) error {
	var companyID *string
	if s.ctx.Request.Header.Get("ID-Company") != "" {
		compID := s.ctx.Request.Header.Get("ID-Company")
		companyID = &compID
	}

	// GET INVENTORY ACCOUNT
	var inventoryAccount models.AccountModel
	err := s.db.Where("is_inventory_account = ? and company_id = ?", true, *companyID).First(&inventoryAccount).Error
	if err != nil {
		return errors.New("inventory account not found")
	}

	return s.db.Create(data).Error
}

// ReceivePurchaseOrder menerima barang dari supplier dan menambah stok
func (s *PurchaseService) ReceivePurchaseOrder(date time.Time, poID, warehouseID string, description string) error {
	// companyID := s.ctx.Request.Header.Get("ID-Company")
	var po models.PurchaseOrderModel
	if err := s.db.First(&po, poID).Error; err != nil {
		return err
	}

	// Pastikan status PO adalah "pending"
	if po.StockStatus != "pending" {
		return errors.New("purchase order already processed")
	}

	err := s.ctx.DB.Transaction(func(tx *gorm.DB) error {
		// do some database operations in the transaction (use 'tx' from this point, not 'db')
		for _, v := range po.Items {
			if v.ProductID == nil || v.WarehouseID == nil {
				continue
			}
			if _, err := s.stockMovementService.AddMovement(date, *v.ProductID, *v.WarehouseID, v.VariantID, nil, nil, nil, v.Quantity, models.MovementTypeIn, po.ID, description); err != nil {
				tx.Rollback()
				return err
			}
		}

		// Update status PO menjadi "received"
		po.StockStatus = "received"
		if err := tx.Save(&po).Error; err != nil {
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

// CancelPurchaseOrder membatalkan purchase order
func (s *PurchaseService) CancelPurchaseOrder(poID uint) error {
	var po models.PurchaseOrderModel
	if err := s.db.First(&po, poID).Error; err != nil {
		return err
	}

	// Pastikan status PO adalah "pending"
	if po.Status != "pending" {
		return errors.New("purchase order already processed")
	}

	// Update status PO menjadi "cancelled"
	po.Status = "cancelled"
	if err := s.db.Save(&po).Error; err != nil {
		return err
	}

	return nil
}

// CreatePayment membuat payment untuk purchase order
func (s *PurchaseService) CreatePayment(poID string, date time.Time, amount float64, accountPayableID *string, accountAssetID string) error {
	var companyID *string
	if s.ctx.Request.Header.Get("ID-Company") != "" {
		compID := s.ctx.Request.Header.Get("ID-Company")
		companyID = &compID
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		var data models.PurchaseOrderModel
		if err := s.db.First(&data, poID).Error; err != nil {
			return err
		}

		// Pastikan status PO adalah "pending"
		if data.Status != "pending" {
			return errors.New("purchase order already processed")
		}

		if data.Paid+amount > data.Total {
			return errors.New("amount is greater than total")
		}

		if err := s.financeService.TransactionService.CreateTransaction(&models.TransactionModel{
			Date:               date,
			AccountID:          &accountAssetID,
			Description:        "Pembayaran " + data.PurchaseNumber,
			Notes:              data.Description,
			TransactionRefID:   &data.ID,
			TransactionRefType: "purchase",
			CompanyID:          companyID,
		}, -amount); err != nil {
			return err
		}

		if accountPayableID != nil {
			if err := s.financeService.TransactionService.CreateTransaction(&models.TransactionModel{
				Date:               date,
				AccountID:          accountPayableID,
				Description:        "Pembayaran " + data.PurchaseNumber,
				Notes:              data.Description,
				TransactionRefID:   &data.ID,
				TransactionRefType: "purchase",
				CompanyID:          companyID,
			}, amount); err != nil {
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

func (s *PurchaseService) GetPurchases(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.Preload("Contact")
	if search != "" {
		stmt = stmt.Where("description ILIKE ? OR purchase_number ILIKE ?",
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
	if request.URL.Query().Get("is_published") != "" {
		stmt = stmt.Where("published_at IS NOT NULL")
	}
	stmt = stmt.Model(&models.PurchaseOrderModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.PurchaseOrderModel{})
	page.Page = page.Page + 1
	return page, nil
}

func (s *PurchaseService) GetPurchaseByID(id string) (*models.PurchaseOrderModel, error) {
	var data models.PurchaseOrderModel
	if err := s.db.Preload("PublishedBy", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, full_name")
	}).Preload("PaymentAccount").Preload("PurchasePayments").First(&data, "id = ?", id).Error; err != nil {
		return nil, err
	}

	if data.RefID != nil {
		var refData models.PurchaseOrderModel
		if err := s.db.First(&refData, "id = ?", *data.RefID).Error; err != nil {
			return nil, err
		}
		data.PurchaseRef = &refData
	}

	paid := 0.0
	for _, v := range data.PurchasePayments {
		paid += v.Amount
	}
	// utils.LogJson(data.PaymentAccount)
	if data.PaymentAccount != nil {
		if data.PaymentAccount.Type == "ASSET" {
			paid = data.Total
		}
	}
	data.Paid = paid
	s.db.Model(&data).Where("id = ?", id).Update("paid", paid)

	return &data, nil
}

func (s *PurchaseService) AddItem(purchase *models.PurchaseOrderModel, data *models.PurchaseOrderItemModel) error {
	if err := s.db.Create(data).Error; err != nil {
		return err
	}
	return s.UpdateTotal(purchase)
}

func (s *PurchaseService) UpdateTotal(purchase *models.PurchaseOrderModel) error {
	s.db.Preload("Items").Model(purchase).Find(purchase)
	var totalBeforeTax, totalBeforeDisc, subTotal, itemsTax, totalDisc float64
	for _, v := range purchase.Items {
		totalBeforeDisc += v.SubtotalBeforeDisc
		totalBeforeTax += v.SubTotal
		subTotal += v.SubTotal
		itemsTax += v.TotalTax
		totalDisc += v.DiscountAmount
	}
	purchase.TotalBeforeTax = totalBeforeTax
	purchase.TotalBeforeDisc = totalBeforeDisc
	purchase.Subtotal = subTotal

	afterTax, purchaseTaxAmount, taxBreakdown := s.CalculateTaxes(subTotal, purchase.IsCompound, purchase.Taxes)
	// fmt.Printf("TAX_AMOUNT %f", purchaseTaxAmount)
	purchase.Subtotal = afterTax
	purchase.TotalTax = itemsTax + purchaseTaxAmount
	purchase.Total = purchase.Subtotal + purchase.TotalTax
	purchase.TotalDiscount = totalDisc
	b, _ := json.Marshal(taxBreakdown)
	purchase.TaxBreakdown = string(b)

	return s.db.Omit(clause.Associations).Save(&purchase).Error
}

func (s *PurchaseService) CalculateTaxes(baseAmount float64, isCompound bool, taxes []*models.TaxModel) (float64, float64, map[string]float64) {
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

func (s *PurchaseService) GetItems(id string) ([]models.PurchaseOrderItemModel, error) {
	var items []models.PurchaseOrderItemModel

	err := s.db.
		Preload("Product", func(db *gorm.DB) *gorm.DB {
			return db.Preload("Category")
		}).
		Preload("Unit").
		Preload("Variant").
		Preload("Warehouse").
		Preload("Tax").
		Where("purchase_id = ?", id).Order("created_at ASC").Find(&items).Error
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

func (s *PurchaseService) DeleteItem(purchase *models.PurchaseOrderModel, itemID string) error {
	err := s.db.Where("purchase_id = ? AND id = ?", purchase.ID, itemID).Delete(&models.PurchaseOrderItemModel{}).Error
	if err != nil {
		return err
	}
	return s.UpdateTotal(purchase)
}

func (s *PurchaseService) UpdateItem(purchase *models.PurchaseOrderModel, itemID string, item *models.PurchaseOrderItemModel) error {
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
	err := s.db.Where("purchase_id = ? AND id = ?", purchase.ID, itemID).Omit("purchase_id").Save(item).Error
	if err != nil {
		return err
	}
	return s.UpdateTotal(purchase)
}

func (s *PurchaseService) PostPurchase(id string, data *models.PurchaseOrderModel, userID string, date time.Time) error {

	if data.DocumentType != "BILL" {
		return errors.New("document type is not bill")
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
	refType := "purchase"
	secRefType := "purchase_item"

	// GET INVENTORY ACCOUNT
	var inventoryAccount models.AccountModel
	err := s.db.Where("is_inventory_account = ? and company_id = ?", true, *data.CompanyID).First(&inventoryAccount).Error
	if err != nil {
		return errors.New("inventory account not found")
	}

	if data.PaymentAccountID == nil {
		return errors.New("payment account is required")
	}
	assetID := utils.Uuid()
	err = s.db.Transaction(func(tx *gorm.DB) error {
		s.financeService.TransactionService.SetDB(tx)
		s.stockMovementService.SetDB(tx)
		totalPayment := 0.0
		for _, v := range data.Items {
			var label = "Pembelian "
			if v.IsCost {
				label = "Biaya "
			}
			err := s.financeService.TransactionService.CreateTransaction(&models.TransactionModel{
				Date:                        date,
				AccountID:                   &inventoryAccount.ID,
				Description:                 label + data.PurchaseNumber,
				Notes:                       v.Description,
				TransactionRefID:            &assetID,
				TransactionRefType:          "transaction",
				TransactionSecondaryRefID:   &data.ID,
				TransactionSecondaryRefType: refType,
				CompanyID:                   data.CompanyID,
				Debit:                       v.SubTotal,
				UserID:                      &userID,
				IsPurchaseCost:              v.IsCost,
				IsPurchase:                  true,
			}, v.SubTotal)
			if err != nil {
				return err
			}

			totalPayment += v.SubTotal + v.TotalTax

			if v.TaxID != nil {
				if v.Tax == nil {
					return errors.New("tax is required")
				}
				if v.Tax.AccountReceivableID == nil {
					return errors.New("tax account receivable ID is required")
				}
				// PIUTANG PAJAK
				err := s.financeService.TransactionService.CreateTransaction(&models.TransactionModel{
					Date:                        date,
					AccountID:                   v.Tax.AccountReceivableID,
					Description:                 "Piutang Pajak " + data.PurchaseNumber,
					Notes:                       v.Description,
					TransactionRefID:            &data.ID,
					TransactionRefType:          refType,
					TransactionSecondaryRefID:   &v.ID,
					TransactionSecondaryRefType: secRefType,
					CompanyID:                   data.CompanyID,
					Debit:                       v.TotalTax,
					UserID:                      &userID,
					IsAccountReceivable:         true,
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
				movement, err := s.stockMovementService.AddMovement(
					time.Now(),
					*v.ProductID,
					*v.WarehouseID,
					v.VariantID,
					nil,
					nil,
					data.CompanyID,
					v.Quantity,
					models.MovementTypePurchase,
					data.ID,
					fmt.Sprintf("Purchase %s (%s)", data.PurchaseNumber, v.Description))
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

			}
			err = tx.Save(v).Error
			if err != nil {
				return err
			}
		}

		if err != nil {
			return err
		}

		err = s.financeService.TransactionService.CreateTransaction(&models.TransactionModel{
			BaseModel:          shared.BaseModel{ID: assetID},
			Date:               date,
			AccountID:          data.PaymentAccountID,
			Description:        "Pembelian " + data.PurchaseNumber,
			TransactionRefID:   &data.ID,
			TransactionRefType: refType,
			CompanyID:          data.CompanyID,
			Credit:             totalPayment,
			UserID:             &userID,
		}, totalPayment)

		return tx.Save(data).Error
	})
	s.financeService.TransactionService.SetDB(s.db)
	s.stockMovementService.SetDB(s.db)
	return err
}

func (s *PurchaseService) GetBalance(purchase *models.PurchaseOrderModel) (float64, error) {
	if purchase.PaymentAccount.Type == "ASSET" {
		return 0, nil
	}
	var amount struct {
		Sum float64 `sql:"sum"`
	}
	err := s.db.Model(&models.PurchasePaymentModel{}).Where("purchase_id = ?", purchase.ID).Select("sum(amount)").Scan(&amount).Error
	if err != nil {
		return 0, err
	}
	if purchase.Total > amount.Sum {
		return purchase.Total - amount.Sum, nil
	}
	return 0, errors.New("payment is more than total")
}
func (s *PurchaseService) CreatePurchasePayment(purchase *models.PurchaseOrderModel, purchasePayment *models.PurchasePaymentModel) error {

	err := s.db.Transaction(func(tx *gorm.DB) error {
		s.financeService.TransactionService.SetDB(tx)
		balance, err := s.GetBalance(purchase)
		if err != nil {
			return err
		}
		fmt.Println("balance", balance)
		fmt.Println("purchasePayment.Amount", purchasePayment.Amount)
		if balance < purchasePayment.Amount {
			return errors.New("payment is more than balance")
		}

		if purchasePayment.AssetAccountID == nil {
			return errors.New("asset account is required")
		}
		if purchase.PaymentAccountID == nil {
			return errors.New("purchase payment account not found")
		}

		if purchase.PaymentAccount.Type != "LIABILITY" {
			return errors.New("purchase payment account type must be LIABILITY")
		}
		paymentAmount := purchasePayment.Amount
		discountAmount := 0.0
		if purchasePayment.PaymentDiscount > 0 {
			paymentAmount = purchasePayment.Amount - (purchasePayment.Amount * (purchasePayment.PaymentDiscount / 100))
			discountAmount = purchasePayment.Amount * (purchasePayment.PaymentDiscount / 100)
		}

		paymentID := uuid.New().String()
		receivableID := uuid.New().String()
		assetTransID := uuid.New().String()

		receivableData := models.TransactionModel{
			Code:                        utils.RandString(10, false),
			BaseModel:                   shared.BaseModel{ID: receivableID},
			Date:                        purchasePayment.PaymentDate,
			AccountID:                   purchase.PaymentAccountID,
			Description:                 "Pembayaran " + purchase.PurchaseNumber,
			Notes:                       purchasePayment.Notes,
			TransactionRefID:            &assetTransID,
			TransactionRefType:          "transaction",
			CompanyID:                   purchase.CompanyID,
			Debit:                       purchasePayment.Amount,
			Amount:                      purchasePayment.Amount,
			UserID:                      purchasePayment.UserID,
			TransactionSecondaryRefID:   &purchase.ID,
			TransactionSecondaryRefType: "purchase",
		}
		receivableData.ID = receivableID
		err = s.db.Create(&receivableData).Error
		if err != nil {
			return err
		}

		assetData := models.TransactionModel{
			Code:                        utils.RandString(10, false),
			BaseModel:                   shared.BaseModel{ID: assetTransID},
			Date:                        purchasePayment.PaymentDate,
			AccountID:                   purchasePayment.AssetAccountID,
			Description:                 "Pembayaran " + purchase.PurchaseNumber,
			Notes:                       purchasePayment.Notes,
			TransactionRefID:            &receivableData.ID,
			TransactionRefType:          "transaction",
			CompanyID:                   purchase.CompanyID,
			Credit:                      paymentAmount,
			Amount:                      paymentAmount,
			UserID:                      purchasePayment.UserID,
			TransactionSecondaryRefID:   &purchase.ID,
			TransactionSecondaryRefType: "purchase",
		}

		assetData.ID = assetTransID
		err = s.db.Create(&assetData).Error
		if err != nil {
			return err
		}

		if discountAmount > 0 {
			var inventoryAccount models.AccountModel
			err := s.db.Where("is_inventory_account = ? and company_id = ?", true, *purchase.CompanyID).First(&inventoryAccount).Error
			if err != nil {
				return errors.New("inventory account not found")
			}
			err = s.db.Create(&models.TransactionModel{
				Code:                        utils.RandString(10, false),
				Date:                        purchasePayment.PaymentDate,
				AccountID:                   &inventoryAccount.ID,
				Description:                 "Diskon " + purchase.PurchaseNumber,
				TransactionRefID:            &receivableData.ID,
				TransactionRefType:          "transaction",
				CompanyID:                   purchase.CompanyID,
				Credit:                      discountAmount,
				Amount:                      discountAmount,
				UserID:                      purchasePayment.UserID,
				TransactionSecondaryRefID:   &purchase.ID,
				TransactionSecondaryRefType: "purchase",
				IsDiscount:                  true,
				Notes:                       purchasePayment.Notes,
			}).Error
			if err != nil {
				return err
			}
		}

		purchasePayment.ID = paymentID

		return tx.Create(purchasePayment).Error
	})
	s.financeService.TransactionService.SetDB(s.db)
	return err
}
