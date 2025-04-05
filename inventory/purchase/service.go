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
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
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

func NewPurchaseService(db *gorm.DB, ctx *context.ERPContext, financeService *finance.FinanceService, stockMovementService *stockmovement.StockMovementService) *PurchaseService {
	return &PurchaseService{
		db:                   db,
		ctx:                  ctx,
		financeService:       financeService,
		stockMovementService: stockMovementService,
	}
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&models.PurchaseOrderModel{}, &models.PurchaseOrderItemModel{})
}

func (s *PurchaseService) UpdatePurchase(id string, data *models.PurchaseOrderModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

func (s *PurchaseService) DeletePurchase(id string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Where("transaction_ref_id = ?", id).Delete(&models.TransactionModel{}).Error
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

// CreatePurchaseOrder membuat purchase order baru
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
		return errors.New("supply account not found")
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
	stmt = stmt.Model(&models.PurchaseOrderModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.PurchaseOrderModel{})
	page.Page = page.Page + 1
	return page, nil
}

func (s *PurchaseService) GetPurchaseByID(id string) (*models.PurchaseOrderModel, error) {
	var data models.PurchaseOrderModel
	if err := s.db.Preload("Company").First(&data, "id = ?", id).Error; err != nil {
		return nil, err
	}

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
	return s.db.Transaction(func(tx *gorm.DB) error {
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
				TransactionRefID:            &data.ID,
				TransactionRefType:          refType,
				TransactionSecondaryRefID:   &v.ID,
				TransactionSecondaryRefType: secRefType,
				CompanyID:                   data.CompanyID,
				Debit:                       v.SubTotal,
				UserID:                      &userID,
				IsPurchaseCost:              v.IsCost,
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
		}

		if err != nil {
			return err
		}

		err = s.financeService.TransactionService.CreateTransaction(&models.TransactionModel{
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
}
