package purchase_return

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/finance"
	"github.com/AMETORY/ametory-erp-modules/inventory/purchase"
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
	purchaseService      *purchase.PurchaseService
}

// NewPurchaseReturnService creates a new instance of PurchaseReturnService with the given database connection, context, finance service, stock movement service and purchase service.
func NewPurchaseReturnService(db *gorm.DB, ctx *context.ERPContext, financeService *finance.FinanceService, stockMovementService *stockmovement.StockMovementService, purchaseService *purchase.PurchaseService) *PurchaseReturnService {
	return &PurchaseReturnService{
		db:                   db,
		ctx:                  ctx,
		financeService:       financeService,
		stockMovementService: stockMovementService,
		purchaseService:      purchaseService,
	}
}

// Migrate migrates the database schema needed for the PurchaseReturnService.
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&models.ReturnModel{}, &models.ReturnItemModel{})
}

// GetReturns retrieves a paginated list of purchase returns from the database.
//
// It takes an http.Request and a search query string as input. The method uses
// GORM to query the database for purchase returns, applying the search query
// to the return description and return number fields. If the request contains
// a company ID header, the method also filters the result by the company ID.
// The function utilizes pagination to manage the result set and applies any
// necessary request modifications using the utils.FixRequest utility.
//
// The function returns a paginated page of ReturnModel and an error if
// the operation fails.
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

// GetReturnByID retrieves a purchase return by its ID.
//
// The function takes the ID of the return as a string and returns a pointer to a ReturnModel
// containing the return details. It preloads the ReleasedBy and Items associations,
// along with related data such as Product, Variant, Unit, Tax, and Warehouse for each item.
// It also fetches the associated purchase order using the RefID and stores it in the PurchaseRef field.
// The function returns an error if the retrieval operation fails.
func (s *PurchaseReturnService) GetReturnByID(id string) (*models.ReturnModel, error) {
	var returnPurchase models.ReturnModel
	err := s.db.Preload("ReleasedBy", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "full_name")
	}).Where("id = ?", id).Preload("Items", func(tx *gorm.DB) *gorm.DB {
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

// AddItem adds a new item to the purchase return with the given ID.
//
// The function takes a pointer to a ReturnModel which contains the return details
// and a pointer to a ReturnItemModel which contains the item details.
// It creates a new record in the return items table with the given data.
// The function returns an error if the operation fails.
func (s *PurchaseReturnService) AddItem(returnPurchase *models.ReturnModel, item *models.ReturnItemModel) error {
	item.ReturnID = returnPurchase.ID
	if err := s.db.Create(item).Error; err != nil {
		return err
	}
	return nil
}

// DeleteItem deletes a purchase return item with the given ID from the database.
//
// It takes the ID of the return and the ID of the item to delete as input.
// The function returns an error if the operation fails.
func (s *PurchaseReturnService) DeleteItem(returnID string, itemID string) error {
	if err := s.db.Where("id = ? AND return_id = ?", itemID, returnID).Delete(&models.ReturnItemModel{}).Error; err != nil {
		return err
	}
	return nil
}

// DeleteReturn removes a purchase return and its associated items from the database.
//
// The function takes the ID of the purchase return as input. It first deletes all items
// associated with the purchase return by their return ID, and then deletes the purchase
// return itself. If any deletion operation fails, the function returns an error.
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

// UpdateReturn updates an existing purchase return with the given ID and data.
//
// The function takes the ID of the purchase return as a string and a pointer to a ReturnModel
// containing the new data. It updates the purchase return record in the database with
// the given values, omitting any associations. If the operation fails, the function returns
// an error.
func (s *PurchaseReturnService) UpdateReturn(id string, returnPurchase *models.ReturnModel) error {
	return s.db.Omit(clause.Associations).Where("id = ?", id).Save(returnPurchase).Error
}

// CreateReturn creates a new purchase return from the given data and commits the transaction.
//
// The function takes a pointer to a ReturnModel as input, which contains the new data.
// It creates a new record in the returns table with the given data, omitting any associations.
// It also creates new records in the return items table for each item in the associated
// purchase order, copying the data from the purchase order items. If the operation fails,
// the function returns an error.
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
			ReturnID:           returnPurchase.ID,
			Description:        v.Description,
			ProductID:          v.ProductID,
			VariantID:          v.VariantID,
			Quantity:           v.Quantity,
			OriginalQuantity:   v.Quantity,
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

// ReleaseReturn processes the release of a purchase return with the specified ID.
//
// This function retrieves the purchase return details, checks for the existence
// of return items, and verifies the necessary accounts. It then performs a series
// of transactions to update financial and stock records, including creating transactions
// for inventory, cash/credit, and tax accounts. Stock movements are recorded for
// each return item, and the associated purchase order is updated. If an account ID
// is provided, additional transactions are performed for asset and source accounts,
// and a purchase payment return is created if applicable.
//
// Parameters:
//   - returnID: The ID of the purchase return to be released.
//   - userID: The ID of the user performing the release operation.
//   - date: The date of the release transaction.
//   - notes: Notes or comments regarding the release.
//   - accountID: An optional account ID used for specific transactions.
//
// Returns an error if any operation within the transaction fails or if required
// data is missing.
func (s *PurchaseReturnService) ReleaseReturn(returnID string, userID string, date time.Time, notes string, accountID *string) error {
	returnPurchase, err := s.GetReturnByID(returnID)
	if err != nil {
		return err
	}
	now := time.Now()

	purchase, err := s.purchaseService.GetPurchaseByID(returnPurchase.RefID)
	if err != nil {
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

	returnTotal := 0.0

	err = s.db.Transaction(func(tx *gorm.DB) error {
		s.financeService.TransactionService.SetDB(tx)
		s.stockMovementService.SetDB(tx)
		for _, v := range returnPurchase.Items {
			returnTotal += v.Total
			assetID := utils.Uuid()
			inventoryID := utils.Uuid()
			// fmt.Println(v)

			// PERSEDIAAN
			err = tx.Create(&models.TransactionModel{
				BaseModel:                   shared.BaseModel{ID: inventoryID},
				Code:                        utils.RandString(10, false),
				Date:                        date,
				AccountID:                   &inventoryAccount.ID,
				Description:                 fmt.Sprintf("[Retur %s] %s", returnPurchase.ReturnNumber, v.Description),
				TransactionRefID:            &assetID,
				TransactionRefType:          "transaction",
				CompanyID:                   purchase.CompanyID,
				Credit:                      v.SubTotal,
				Amount:                      v.SubTotal,
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
			err = tx.Create(&models.TransactionModel{
				BaseModel:                   shared.BaseModel{ID: assetID},
				Code:                        utils.RandString(10, false),
				Date:                        date,
				AccountID:                   purchase.PaymentAccountID,
				Description:                 fmt.Sprintf("[Retur %s] %s", returnPurchase.ReturnNumber, v.Description),
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
				err = tx.Create(&models.TransactionModel{
					BaseModel:                   shared.BaseModel{ID: returnCreditID},
					Code:                        utils.RandString(10, false),
					Date:                        date,
					AccountID:                   purchase.PaymentAccountID,
					Description:                 fmt.Sprintf("[Retur %s] %s", returnPurchase.ReturnNumber, v.Description),
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

			// UPDATE INVOICE
			purchaseItem := models.PurchaseOrderItemModel{
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
				PurchaseID:      &purchase.ID,
			}
			err = s.purchaseService.AddItem(purchase, &purchaseItem)
			if err != nil {
				return err
			}

			err = s.purchaseService.UpdateItem(purchase, purchaseItem.ID, &purchaseItem)
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
				err := s.financeService.TransactionService.CreateTransaction(&models.TransactionModel{
					Date:                        date,
					AccountID:                   v.Tax.AccountReceivableID,
					Description:                 "Retur Piutang Pajak " + returnPurchase.ReturnNumber,
					Notes:                       v.Description,
					TransactionRefID:            &returnPurchase.ID,
					TransactionRefType:          returnRefType,
					TransactionSecondaryRefID:   &purchase.ID,
					TransactionSecondaryRefType: returnSecRefType,
					CompanyID:                   returnPurchase.CompanyID,
					Credit:                      v.TotalTax,
					UserID:                      &userID,
					IsAccountReceivable:         true,
					IsTax:                       true,
				}, v.TotalTax)
				if err != nil {
					return err
				}

			}
		}
		// Commit the transaction
		returnPurchase.Status = "RELEASED"
		returnPurchase.ReleasedAt = &now
		returnPurchase.ReleasedByID = &userID
		// CLEAR TRANSACTION
		s.purchaseService.UpdateTotal(purchase)

		if accountID != nil {
			// if accountID is ASSET, CREATE purcahase payment return
			var account models.AccountModel
			err = tx.Model(&account).Where("id = ?", accountID).First(&account).Error
			if err != nil {
				return err
			}

			if account.Type == models.ASSET {
				if purchase.Paid < returnTotal {
					return errors.New("paid is less than return total")
				}
				tx.Create(&models.PurchasePaymentModel{
					PurchaseID:  &purchase.ID,
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

// UpdateItem updates the details of a return item in the database.
//
// This function recalculates and updates the value, subtotal before discount,
// discount amount, subtotal, total tax, and total for the given return item
// based on its quantity, unit price, discount percent, and tax information.
// It fetches the unit value if a UnitID is provided and updates the item value
// accordingly. The function also saves the updated item in the database.
//
// Parameters:
//   - item: A pointer to a ReturnItemModel containing the details of the return
//     item to be updated.
//
// Returns:
//   - An error if the database update operation fails; otherwise, nil.
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
	item.TotalTax = taxAmount
	item.SubtotalBeforeDisc = subtotalBeforeDisc

	return s.db.Save(item).Error
}
