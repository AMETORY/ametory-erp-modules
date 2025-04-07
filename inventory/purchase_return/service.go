package purchase_return

import (
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
	"github.com/morkid/paginate"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PurchaseReturnService struct {
	db                   *gorm.DB
	ctx                  *context.ERPContext
	financeService       *finance.FinanceService
	stockMovementService *stockmovement.StockMovementService
}

func NewPurchaseReturnService(db *gorm.DB, ctx *context.ERPContext, financeService *finance.FinanceService, stockMovementService *stockmovement.StockMovementService) *PurchaseReturnService {
	return &PurchaseReturnService{
		db:                   db,
		ctx:                  ctx,
		financeService:       financeService,
		stockMovementService: stockMovementService,
	}
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&models.ReturnModel{}, &models.ReturnItemModel{})
}

func (s *PurchaseReturnService) GetReturns(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.Preload("Items")
	if search != "" {
		stmt = stmt.Where("description ILIKE ? OR return_number ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
		)
	}
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	stmt = stmt.Where("return_type = ?", "PURCHASE_RETURN")
	stmt = stmt.Model(&models.ReturnModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.ReturnModel{})
	page.Page = page.Page + 1

	items := page.Items.(*[]models.ReturnModel)
	newItems := make([]models.ReturnModel, 0) // TODO: optimize
	for _, item := range *items {
		var purchase models.PurchaseOrderModel
		if err := s.db.Model(&models.PurchaseOrderModel{}).Where("id = ?", item.RefID).First(&purchase).Error; err != nil {
			return page, err
		}
		item.PurchaseRef = &purchase
		newItems = append(newItems, item)
	}
	page.Items = &newItems
	return page, nil
}

func (s *PurchaseReturnService) GetReturnByID(id string) (*models.ReturnModel, error) {
	var returnPurchase models.ReturnModel
	err := s.db.Where("id = ?", id).Preload("Items", func(tx *gorm.DB) *gorm.DB {
		return tx.Preload("Product").Preload("Variant").Preload("Unit").Preload("Tax").Preload("Warehouse")
	}).First(&returnPurchase).Error
	if err != nil {
		return nil, err
	}
	var purchase models.PurchaseOrderModel
	if err := s.db.Model(&models.PurchaseOrderModel{}).Preload("PaymentAccount").Where("id = ?", returnPurchase.RefID).First(&purchase).Error; err != nil {
		return nil, err
	}
	returnPurchase.PurchaseRef = &purchase
	return &returnPurchase, nil
}
func (s *PurchaseReturnService) AddItem(returnPurchase *models.ReturnModel, item *models.ReturnItemModel) error {
	item.ReturnID = returnPurchase.ID
	if err := s.db.Create(item).Error; err != nil {
		return err
	}
	return nil
}

func (s *PurchaseReturnService) DeleteItem(returnID string, itemID string) error {
	if err := s.db.Where("id = ? AND return_id = ?", itemID, returnID).Delete(&models.ReturnItemModel{}).Error; err != nil {
		return err
	}
	return nil
}

func (s *PurchaseReturnService) DeleteReturn(id string) error {
	// Delete all items first
	if err := s.db.Where("return_id = ?", id).Delete(&models.ReturnItemModel{}).Error; err != nil {
		return err
	}
	if err := s.db.Where("id = ?", id).Delete(&models.ReturnModel{}).Error; err != nil {
		return err
	}
	return nil
}

func (s *PurchaseReturnService) UpdateReturn(id string, returnPurchase *models.ReturnModel) error {
	return s.db.Omit(clause.Associations).Where("id = ?", id).Save(returnPurchase).Error
}
func (s *PurchaseReturnService) CreateReturn(returnPurchase *models.ReturnModel) error {
	returnPurchase.ReturnType = "PURCHASE_RETURN"
	// Commit the transaction

	var purchaseItems []models.PurchaseOrderItemModel
	s.db.Where("purchase_id = ?", returnPurchase.RefID).Find(&purchaseItems)
	var items []models.ReturnItemModel
	for _, v := range purchaseItems {
		if v.ProductID == nil {
			continue
		}
		items = append(items, models.ReturnItemModel{
			ReturnID:         returnPurchase.ID,
			Description:      v.Description,
			ProductID:        v.ProductID,
			VariantID:        v.VariantID,
			Quantity:         v.Quantity,
			OriginalQuantity: v.Quantity,
			UnitPrice:        v.UnitPrice,
			UnitID:           v.UnitID,
			Value:            v.UnitValue,
			Total:            v.Total,
			SubTotal:         v.SubTotal,
			DiscountPercent:  v.DiscountPercent,
			DiscountAmount:   v.DiscountAmount,
			TaxID:            v.TaxID,
			WarehouseID:      v.WarehouseID,
		})
	}
	returnPurchase.Items = items
	return s.db.Create(returnPurchase).Error
}

func (s *PurchaseReturnService) ReleaseReturn(returnID string, userID string, date time.Time, notes string, accountID *string) error {
	returnPurchase, err := s.GetReturnByID(returnID)
	if err != nil {
		return err
	}
	now := time.Now()

	var purchase models.PurchaseOrderModel
	if err := s.db.Where("id = ?", returnPurchase.RefID).First(&purchase).Error; err != nil {
		return err
	}

	if len(returnPurchase.Items) == 0 {
		return errors.New("return items is empty")
	}
	var inventoryAccount models.AccountModel
	err = s.db.Where("is_inventory_account = ? and company_id = ?", true, *purchase.CompanyID).First(&inventoryAccount).Error
	if err != nil {
		return errors.New("inventory account not found")
	}

	returnRefType := "return_purchase"
	returnSecRefType := "purchase"

	for _, v := range returnPurchase.Items {

		assetID := utils.Uuid()
		inventoryID := utils.Uuid()
		// fmt.Println(v)

		// PERSEDIAAN
		err = s.db.Create(&models.TransactionModel{
			BaseModel:                   shared.BaseModel{ID: inventoryID},
			Code:                        utils.RandString(10, false),
			Date:                        date,
			AccountID:                   &inventoryAccount.ID,
			Description:                 "Retur " + purchase.PurchaseNumber,
			TransactionRefID:            &assetID,
			TransactionRefType:          "transaction",
			CompanyID:                   purchase.CompanyID,
			Credit:                      v.Total,
			Amount:                      v.Total,
			UserID:                      &userID,
			TransactionSecondaryRefID:   &returnID,
			TransactionSecondaryRefType: "return_purchase",
			IsReturn:                    true,
			Notes:                       returnPurchase.Notes,
		}).Error
		if err != nil {
			return err
		}

		// CASH / HUTANG
		err = s.db.Create(&models.TransactionModel{
			BaseModel:                   shared.BaseModel{ID: assetID},
			Code:                        utils.RandString(10, false),
			Date:                        date,
			AccountID:                   purchase.PaymentAccountID,
			Description:                 "Retur " + purchase.PurchaseNumber,
			TransactionRefID:            &inventoryID,
			TransactionRefType:          "transaction",
			CompanyID:                   purchase.CompanyID,
			Debit:                       v.Total,
			Amount:                      v.Total,
			UserID:                      &userID,
			TransactionSecondaryRefID:   &returnID,
			TransactionSecondaryRefType: "return_purchase",
			IsReturn:                    true,
			Notes:                       returnPurchase.Notes,
		}).Error
		if err != nil {
			return err
		}

		// STOCK MOVEMENT

		movement, err := s.stockMovementService.AddMovement(
			time.Now(),
			*v.ProductID,
			*v.WarehouseID,
			v.VariantID,
			nil,
			nil,
			returnPurchase.CompanyID,
			-v.Quantity,
			models.MovementTypeReturn,
			returnID,
			fmt.Sprintf("Return %s (%s)", returnPurchase.ReturnNumber, v.Description))
		if err != nil {
			return err
		}
		movement.ReferenceID = returnID
		movement.ReferenceType = &returnRefType
		movement.SecondaryRefID = &purchase.ID
		movement.SecondaryRefType = &returnSecRefType
		movement.Value = v.Value
		movement.UnitID = v.UnitID

		err = s.db.Save(movement).Error
		if err != nil {
			return err
		}

		if accountID != nil {
			returnAssetID := utils.Uuid()
			returnCreditID := utils.Uuid()

			// RETURN ASSET
			err = s.db.Create(&models.TransactionModel{
				BaseModel:                   shared.BaseModel{ID: returnAssetID},
				Code:                        utils.RandString(10, false),
				Date:                        date,
				AccountID:                   accountID,
				Description:                 "Retur " + purchase.PurchaseNumber,
				TransactionRefID:            &returnCreditID,
				TransactionRefType:          "transaction",
				CompanyID:                   purchase.CompanyID,
				Debit:                       v.Total,
				Amount:                      v.Total,
				UserID:                      &userID,
				TransactionSecondaryRefID:   &returnID,
				TransactionSecondaryRefType: "return_purchase",
				IsReturn:                    true,
				Notes:                       returnPurchase.Notes,
			}).Error
			if err != nil {
				return err
			}

			// SOURCE ACCOUNT
			err = s.db.Create(&models.TransactionModel{
				BaseModel:                   shared.BaseModel{ID: returnCreditID},
				Code:                        utils.RandString(10, false),
				Date:                        date,
				AccountID:                   purchase.PaymentAccountID,
				Description:                 "Retur " + purchase.PurchaseNumber,
				TransactionRefID:            &returnAssetID,
				TransactionRefType:          "transaction",
				CompanyID:                   purchase.CompanyID,
				Credit:                      v.Total,
				Amount:                      v.Total,
				UserID:                      &userID,
				TransactionSecondaryRefID:   &returnID,
				TransactionSecondaryRefType: "return_purchase",
				IsReturn:                    true,
				Notes:                       returnPurchase.Notes,
			}).Error
			if err != nil {
				return err
			}
		}
	}
	// Commit the transaction
	returnPurchase.Status = "RELEASED"
	returnPurchase.ReleasedAt = &now
	returnPurchase.ReleasedByID = &userID

	return s.UpdateReturn(returnID, returnPurchase)
}

func (s *PurchaseReturnService) UpdateItem(item *models.ReturnItemModel) error {
	taxPercent := 0.0
	taxAmount := 0.0

	if item.UnitID != nil {
		productUnit := models.ProductUnitData{}
		s.db.Model(&productUnit).Where("product_model_id = ? and unit_model_id = ?", item.ProductID, item.UnitID).Find(&productUnit)
		item.Value = productUnit.Value
	} else {
		item.Value = 1
	}

	if item.TaxID != nil {
		taxPercent = item.Tax.Amount
	}
	subtotalBeforeDisc := (item.Quantity * item.Value) * item.UnitPrice
	if item.DiscountPercent > 0 {
		taxAmount = (subtotalBeforeDisc - (subtotalBeforeDisc * item.DiscountPercent / 100)) * (taxPercent / 100)
		item.SubTotal = (subtotalBeforeDisc - (subtotalBeforeDisc * item.DiscountPercent / 100))
		item.DiscountAmount = subtotalBeforeDisc * item.DiscountPercent / 100
	} else {
		taxAmount = (subtotalBeforeDisc - item.DiscountAmount) * (taxPercent / 100)
		item.SubTotal = (subtotalBeforeDisc - item.DiscountAmount)
		item.DiscountPercent = 0
	}
	// item.TotalTax = taxAmount
	item.Total = item.SubTotal + taxAmount

	return s.db.Save(item).Error
}
