package sales_return

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/finance"
	stockmovement "github.com/AMETORY/ametory-erp-modules/inventory/stock_movement"
	"github.com/AMETORY/ametory-erp-modules/order/sales"
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SalesReturnService struct {
	db                   *gorm.DB
	ctx                  *context.ERPContext
	financeService       *finance.FinanceService
	stockMovementService *stockmovement.StockMovementService
	salesService         *sales.SalesService
}

func NewSalesReturnService(db *gorm.DB, ctx *context.ERPContext, financeService *finance.FinanceService, stockMovementService *stockmovement.StockMovementService, salesService *sales.SalesService) *SalesReturnService {
	return &SalesReturnService{
		db:                   db,
		ctx:                  ctx,
		stockMovementService: stockMovementService,
		salesService:         salesService,
		financeService:       financeService,
	}
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&models.ReturnModel{}, &models.ReturnItemModel{})
}

func (s *SalesReturnService) GetReturns(request http.Request, search string) (paginate.Page, error) {
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
	stmt = stmt.Where("return_type = ?", "SALES_RETURN")
	stmt = stmt.Model(&models.ReturnModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.ReturnModel{})
	page.Page = page.Page + 1

	items := page.Items.(*[]models.ReturnModel)
	newItems := make([]models.ReturnModel, 0) // TODO: optimize
	for _, item := range *items {
		var sales models.SalesModel
		if err := s.db.Model(&models.SalesModel{}).Where("id = ?", item.RefID).First(&sales).Error; err != nil {
			return page, err
		}
		item.SalesRef = &sales
		newItems = append(newItems, item)
	}
	page.Items = &newItems
	return page, nil
}

func (s *SalesReturnService) GetReturnByID(id string) (*models.ReturnModel, error) {
	var returnPurchase models.ReturnModel
	err := s.db.Preload("ReleasedBy", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "full_name")
	}).Where("id = ?", id).Preload("Items", func(tx *gorm.DB) *gorm.DB {
		return tx.Preload("Product").Preload("Variant").Preload("Unit").Preload("Tax").Preload("Warehouse")
	}).First(&returnPurchase).Error
	if err != nil {
		return nil, err
	}
	var sales models.SalesModel
	if err := s.db.Model(&models.SalesModel{}).Preload("PaymentAccount").Where("id = ?", returnPurchase.RefID).First(&sales).Error; err != nil {
		return nil, err
	}
	returnPurchase.SalesRef = &sales
	return &returnPurchase, nil
}
func (s *SalesReturnService) AddItem(returnPurchase *models.ReturnModel, item *models.ReturnItemModel) error {
	item.ReturnID = returnPurchase.ID
	if err := s.db.Create(item).Error; err != nil {
		return err
	}
	return nil
}

func (s *SalesReturnService) DeleteItem(returnID string, itemID string) error {
	if err := s.db.Where("id = ? AND return_id = ?", itemID, returnID).Delete(&models.ReturnItemModel{}).Error; err != nil {
		return err
	}
	return nil
}

func (s *SalesReturnService) DeleteReturn(id string) error {
	// Delete all items first
	if err := s.db.Where("return_id = ?", id).Delete(&models.ReturnItemModel{}).Error; err != nil {
		return err
	}
	if err := s.db.Where("id = ?", id).Delete(&models.ReturnModel{}).Error; err != nil {
		return err
	}
	return nil
}

func (s *SalesReturnService) UpdateReturn(id string, returnPurchase *models.ReturnModel) error {
	return s.db.Omit(clause.Associations).Where("id = ?", id).Save(returnPurchase).Error
}
func (s *SalesReturnService) CreateReturn(returnPurchase *models.ReturnModel) error {
	returnPurchase.ReturnType = "SALES_RETURN"
	// Commit the transaction

	var salesItems []models.SalesItemModel
	s.db.Where("sales_id = ?", returnPurchase.RefID).Find(&salesItems)
	var items []models.ReturnItemModel
	for _, v := range salesItems {
		if v.ProductID == nil {
			continue
		}
		items = append(items, models.ReturnItemModel{
			ReturnID:           returnPurchase.ID,
			Description:        v.Description,
			ProductID:          v.ProductID,
			VariantID:          v.VariantID,
			Quantity:           v.Quantity,
			OriginalQuantity:   v.Quantity,
			BasePrice:          v.BasePrice,
			UnitPrice:          v.UnitPrice,
			UnitID:             v.UnitID,
			Value:              v.UnitValue,
			Total:              v.Total,
			SubTotal:           v.SubTotal,
			SubtotalBeforeDisc: v.SubtotalBeforeDisc,
			TotalTax:           v.TotalTax,
			DiscountPercent:    v.DiscountPercent,
			DiscountAmount:     v.DiscountAmount,
			TaxID:              v.TaxID,
			WarehouseID:        v.WarehouseID,
		})
	}
	returnPurchase.Items = items
	return s.db.Create(returnPurchase).Error
}

func (s *SalesReturnService) ReleaseReturn(returnID string, userID string, date time.Time, notes string, accountID *string) error {
	returnPurchase, err := s.GetReturnByID(returnID)
	if err != nil {
		return err
	}
	now := time.Now()

	sales, err := s.salesService.GetSalesByID(returnPurchase.RefID)
	if err != nil {
		return err
	}

	if len(returnPurchase.Items) == 0 {
		return errors.New("return items is empty")
	}
	var inventoryAccount models.AccountModel
	err = s.db.Where("is_inventory_account = ? and company_id = ?", true, *sales.CompanyID).First(&inventoryAccount).Error
	if err != nil {
		return errors.New("inventory account not found")
	}
	var cogsAccount models.AccountModel
	err = s.db.Where("is_cogs_account = ? and company_id = ?", true, *sales.CompanyID).First(&cogsAccount).Error
	if err != nil {
		return errors.New("cogs account not found")
	}

	var returnAccount models.AccountModel
	err = s.db.Where("type = ? and company_id = ? and is_return = ?", models.CONTRA_REVENUE, sales.CompanyID, true).First(&returnAccount).Error
	if err != nil {
		return errors.New("return account not found")
	}

	returnRefType := "return_sales"
	returnSecRefType := "sales"

	returnTotal := 0.0

	err = s.db.Transaction(func(tx *gorm.DB) error {
		s.financeService.TransactionService.SetDB(tx)
		s.stockMovementService.SetDB(tx)
		for _, v := range returnPurchase.Items {
			returnTotal += v.Total
			assetID := utils.Uuid()
			inventoryID := utils.Uuid()
			returnTransID := utils.Uuid()
			hppID := utils.Uuid()
			// fmt.Println(v)

			// RETUR AKUN
			err = tx.Create(&models.TransactionModel{
				BaseModel:                   shared.BaseModel{ID: returnTransID},
				Code:                        utils.RandString(10, false),
				Date:                        date,
				AccountID:                   &returnAccount.ID,
				Description:                 fmt.Sprintf("[Retur %s] %s", returnPurchase.ReturnNumber, v.Description),
				TransactionRefID:            &assetID,
				TransactionRefType:          "transaction",
				CompanyID:                   sales.CompanyID,
				Debit:                       v.SubTotal,
				Amount:                      v.SubTotal,
				UserID:                      &userID,
				TransactionSecondaryRefID:   &returnID,
				TransactionSecondaryRefType: "return_sales",
				Notes:                       returnPurchase.Notes,
			}).Error
			if err != nil {
				return err
			}

			// UPDATE TAX
			if v.TaxID != nil {
				if v.Tax == nil {
					return errors.New("tax is required")
				}
				if v.Tax.AccountReceivableID == nil {
					return errors.New("tax account receivable ID is required")
				}
				// PIUTANG PAJAK
				err := tx.Create(&models.TransactionModel{
					Date:                        date,
					AccountID:                   v.Tax.AccountPayableID,
					Description:                 "Retur Piutang Pajak " + returnPurchase.ReturnNumber,
					Notes:                       v.Description,
					TransactionRefID:            &returnPurchase.ID,
					TransactionRefType:          returnRefType,
					TransactionSecondaryRefID:   &sales.ID,
					TransactionSecondaryRefType: returnSecRefType,
					CompanyID:                   returnPurchase.CompanyID,
					Debit:                       v.TotalTax,
					Amount:                      v.TotalTax,
					UserID:                      &userID,
					IsAccountPayable:            true,
					IsTax:                       true,
				}).Error
				if err != nil {
					return err
				}

			}

			// CASH / PIUTANG
			err = tx.Create(&models.TransactionModel{
				BaseModel:                   shared.BaseModel{ID: assetID},
				Code:                        utils.RandString(10, false),
				Date:                        date,
				AccountID:                   sales.PaymentAccountID,
				Description:                 fmt.Sprintf("[Retur %s] %s", returnPurchase.ReturnNumber, v.Description),
				TransactionRefID:            &inventoryID,
				TransactionRefType:          "transaction",
				CompanyID:                   sales.CompanyID,
				Credit:                      v.Total,
				Amount:                      v.Total,
				UserID:                      &userID,
				TransactionSecondaryRefID:   &returnID,
				TransactionSecondaryRefType: "return_sales",
				Notes:                       returnPurchase.Notes,
			}).Error
			if err != nil {
				return err
			}

			// PERSEDIAAN
			err = tx.Create(&models.TransactionModel{
				BaseModel:                   shared.BaseModel{ID: inventoryID},
				Code:                        utils.RandString(10, false),
				Date:                        date,
				AccountID:                   &inventoryAccount.ID,
				Description:                 fmt.Sprintf("[Retur %s] %s", returnPurchase.ReturnNumber, v.Description),
				TransactionRefID:            &hppID,
				TransactionRefType:          "transaction",
				CompanyID:                   sales.CompanyID,
				Debit:                       v.BasePrice * v.Quantity * v.Value,
				Amount:                      v.BasePrice * v.Quantity * v.Value,
				UserID:                      &userID,
				TransactionSecondaryRefID:   &returnID,
				TransactionSecondaryRefType: "return_sales",
				Notes:                       returnPurchase.Notes,
			}).Error
			if err != nil {
				return err
			}

			// HPP
			err = tx.Create(&models.TransactionModel{
				BaseModel:                   shared.BaseModel{ID: hppID},
				Code:                        utils.RandString(10, false),
				Date:                        date,
				AccountID:                   &cogsAccount.ID,
				Description:                 fmt.Sprintf("[Retur %s] %s", returnPurchase.ReturnNumber, v.Description),
				TransactionRefID:            &inventoryID,
				TransactionRefType:          "transaction",
				CompanyID:                   sales.CompanyID,
				Credit:                      v.BasePrice * v.Quantity * v.Value,
				Amount:                      v.BasePrice * v.Quantity * v.Value,
				UserID:                      &userID,
				TransactionSecondaryRefID:   &returnID,
				TransactionSecondaryRefType: "return_sales",
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
				v.Quantity,
				models.MovementTypeReturn,
				returnID,
				fmt.Sprintf("Return %s (%s)", returnPurchase.ReturnNumber, v.Description))
			if err != nil {
				return err
			}
			movement.ReferenceID = returnID
			movement.ReferenceType = &returnRefType
			movement.SecondaryRefID = &sales.ID
			movement.SecondaryRefType = &returnSecRefType
			movement.Value = v.Value
			movement.UnitID = v.UnitID

			err = tx.Save(movement).Error
			if err != nil {
				return err
			}

			if accountID != nil {
				returnAssetID := utils.Uuid()
				returnCreditID := utils.Uuid()

				// RETURN ASSET
				err = tx.Create(&models.TransactionModel{
					BaseModel:                   shared.BaseModel{ID: returnAssetID},
					Code:                        utils.RandString(10, false),
					Date:                        date,
					AccountID:                   accountID,
					Description:                 fmt.Sprintf("[Retur %s] %s", returnPurchase.ReturnNumber, v.Description),
					TransactionRefID:            &returnCreditID,
					TransactionRefType:          "transaction",
					CompanyID:                   sales.CompanyID,
					Credit:                      v.Total,
					Amount:                      v.Total,
					UserID:                      &userID,
					TransactionSecondaryRefID:   &returnID,
					TransactionSecondaryRefType: "return_sales",
					Notes:                       returnPurchase.Notes,
				}).Error
				if err != nil {
					return err
				}

				// SOURCE ACCOUNT
				err = tx.Create(&models.TransactionModel{
					BaseModel:                   shared.BaseModel{ID: returnCreditID},
					Code:                        utils.RandString(10, false),
					Date:                        date,
					AccountID:                   sales.PaymentAccountID,
					Description:                 fmt.Sprintf("[Retur %s] %s", returnPurchase.ReturnNumber, v.Description),
					TransactionRefID:            &returnAssetID,
					TransactionRefType:          "transaction",
					CompanyID:                   sales.CompanyID,
					Debit:                       v.Total,
					Amount:                      v.Total,
					UserID:                      &userID,
					TransactionSecondaryRefID:   &returnID,
					TransactionSecondaryRefType: "return_sales",
					Notes:                       returnPurchase.Notes,
				}).Error
				if err != nil {
					return err
				}
			}

			// UPDATE INVOICE
			salesItem := models.SalesItemModel{
				Description:     fmt.Sprintf("[Retur %s] %s", returnPurchase.ReturnNumber, v.Description),
				Notes:           v.Notes,
				ProductID:       v.ProductID,
				VariantID:       v.VariantID,
				Quantity:        -v.Quantity,
				UnitPrice:       v.UnitPrice,
				UnitID:          v.UnitID,
				UnitValue:       v.Value,
				Total:           v.Total,
				SubTotal:        v.SubTotal,
				DiscountPercent: v.DiscountPercent,
				DiscountAmount:  v.DiscountAmount,
				TaxID:           v.TaxID,
				Tax:             v.Tax,
				WarehouseID:     v.WarehouseID,
				SalesID:         &sales.ID,
			}
			err = s.salesService.AddItem(sales, &salesItem)
			if err != nil {
				return err
			}

			err = s.salesService.UpdateItem(sales, salesItem.ID, &salesItem)
			if err != nil {
				return err
			}

		}
		// Commit the transaction
		returnPurchase.Status = "RELEASED"
		returnPurchase.ReleasedAt = &now
		returnPurchase.ReleasedByID = &userID
		// CLEAR TRANSACTION
		s.salesService.UpdateTotal(sales)

		if accountID != nil {
			// if accountID is ASSET, CREATE purcahase payment return
			var account models.AccountModel
			err = tx.Model(&account).Where("id = ?", accountID).First(&account).Error
			if err != nil {
				return err
			}

			if account.Type == models.ASSET {
				if sales.Paid < returnTotal {
					return errors.New("paid is less than return total")
				}
				tx.Create(&models.SalesPaymentModel{
					SalesID:     &sales.ID,
					PaymentDate: date,
					Amount:      -returnTotal,
					Notes:       fmt.Sprintf("Retur Pembayaran %s", returnPurchase.ReturnNumber),
					UserID:      &userID,
					CompanyID:   returnPurchase.CompanyID,
					IsRefund:    true,
				})
			}
		}
		return s.UpdateReturn(returnID, returnPurchase)
	})

	s.financeService.TransactionService.SetDB(s.db)
	s.stockMovementService.SetDB(s.db)

	return err
}

func (s *SalesReturnService) UpdateItem(item *models.ReturnItemModel) error {
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
	item.TotalTax = taxAmount
	item.SubtotalBeforeDisc = subtotalBeforeDisc

	return s.db.Save(item).Error
}
