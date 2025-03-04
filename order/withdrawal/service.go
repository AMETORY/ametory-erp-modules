package withdrawal

import (
	"errors"
	"net/http"
	"time"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/finance"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type WithdrawalService struct {
	ctx *context.ERPContext
	db  *gorm.DB
}

func Migrate(db *gorm.DB) error {
	// db.Migrator().AlterColumn(&models.WithdrawalItemModel{}, "pos_id")
	db.Migrator().CreateConstraint(&models.WithdrawalModel{}, "pos_id")
	return db.AutoMigrate(&models.WithdrawalModel{}, &models.WithdrawalItemModel{})
}
func NewWithdrawalService(db *gorm.DB, ctx *context.ERPContext) *WithdrawalService {
	return &WithdrawalService{db: db, ctx: ctx}
}

func (w *WithdrawalService) RequestWithdrawal(request *models.WithdrawalModel) (err error) {
	return w.db.Create(request).Error
}

func (w *WithdrawalService) ProcessWithdrawal(withdrawalID string, status models.WithdrawalModel, fileIDs []string) error {
	for _, v := range fileIDs {
		var file models.FileModel
		if err := w.db.Where("id = ?", v).First(&file).Error; err != nil {
			return err
		}
		file.RefID = withdrawalID
		file.RefType = "withdrawal"
		w.db.Save(&file)
	}
	return w.db.Model(&models.WithdrawalModel{}).Where("id = ?", withdrawalID).Update("status", status).Error
}

func (w *WithdrawalService) GetWithdrawal(withdrawalID string) (withdrawal *models.WithdrawalModel, err error) {
	withdrawal = &models.WithdrawalModel{}
	return withdrawal, w.db.Preload("Merchant").
		Preload("ApprovalByUser").
		Preload("RejectedByUser").
		Preload("RequestedByUser").
		Preload("Items", func(db *gorm.DB) *gorm.DB {
			return db.Preload("Pos.Items").Preload("Sales.Items")
		}).Where("id = ?", withdrawalID).First(withdrawal).Error
}
func (w *WithdrawalService) SettlementWithdrawalByID(withdrawalID string, approverID *string) (*models.WithdrawalModel, error) {
	withdrawal := &models.WithdrawalModel{}
	err := w.db.Preload("Merchant").
		Preload("ApprovalByUser").
		Preload("RejectedByUser").
		Preload("RequestedByUser").
		Preload("Items", func(db *gorm.DB) *gorm.DB {
			return db.Preload("Pos.Items").Preload("Sales.Items")
		}).Where("id = ? and status = ?", withdrawalID, "PENDING").First(withdrawal).Error
	if err != nil {
		return nil, err
	}
	now := time.Now()
	withdrawal.Status = "SETTLEMENT"
	withdrawal.DisbursementDate = &now
	withdrawal.ApprovalBy = approverID
	withdrawal.ApprovalDate = &now

	for _, v := range withdrawal.Items {
		if v.Pos != nil {
			v.Pos.Status = "COMPLETED"
			w.db.Save(&v.Pos)
		}
	}

	return withdrawal, w.db.Save(withdrawal).Error
}
func (w *WithdrawalService) GetWithdrawals(request http.Request, search string, merchantID, userID *string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := w.db.
		Preload("Merchant").
		Preload("ApprovalByUser").
		Preload("RejectedByUser").
		Preload("RequestedByUser").
		Preload("Items", func(db *gorm.DB) *gorm.DB {
			return db.Preload("Pos.Items").Preload("Sales.Items")
		}).Model(&models.WithdrawalModel{})
	if search != "" {
		stmt = stmt.Where("withdrawals.code ILIKE ? OR withdrawals.description ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
		)
	}

	if merchantID != nil {
		stmt = stmt.Where("withdrawals.merchant_id = ?", *merchantID)
	}

	if userID != nil {
		stmt = stmt.Where("withdrawals.requested_by = ?", *userID)
	}

	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.WithdrawalModel{})
	page.Page = page.Page + 1
	return page, nil
}
func (w *WithdrawalService) GetOrderWithdrawable(request http.Request, search string, merchantID string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := w.db.Preload("Merchant").Preload("Items", func(tx *gorm.DB) *gorm.DB {
		return tx.Preload("Product").Preload("Variant")
	}).Preload("Payment").
		Joins("JOIN payments ON payments.id = pos_sales.payment_id").
		Joins("LEFT JOIN withdrawal_items ON withdrawal_items.pos_id = pos_sales.id").
		Where("withdrawal_items.pos_id IS NULL").
		Where("merchant_id = ?", merchantID).
		Where("payments.status = ?", "COMPLETE").
		Where("pos_sales.status = ?", "COMPLETED").
		Where("pos_sales.user_payment_status = ?", "PAID").
		Where("payments.payment_method <> ?", "CASH")
	if search != "" {
		stmt = stmt.Where("pos_sales.code ILIKE ? OR pos_sales.description ILIKE ? OR pos_sales.sales_number ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
			"%"+search+"%",
		)
	}
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}

	stmt = stmt.Where("merchant_id = ?", merchantID)

	orderBy := request.URL.Query().Get("order_by")
	order := request.URL.Query().Get("order")
	if orderBy == "" {
		orderBy = "created_at"
	}
	if order == "" {
		order = "desc"
	}
	stmt = stmt.Order(orderBy + " " + order)

	stmt = stmt.Model(&models.POSModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.POSModel{})
	page.Page = page.Page + 1
	// items := page.Items.(*[]models.POSModel)
	// newItems := make([]models.POSModel, 0)
	// for _, item := range *items {
	// 	for _, v := range item.Items {
	// 		images, _ := s.inventoryService.ProductService.ListImagesOfProduct(*v.ProductID)
	// 		v.Product.ProductImages = images
	// 	}

	// 	item.ShippingStatus = "PENDING"

	// 	var shipping models.ShippingModel
	// 	err := s.db.First(&shipping, "order_id = ?", item.ID).Error
	// 	if err == nil {
	// 		item.Shipping = &shipping
	// 		item.ShippingStatus = shipping.Status

	// 	}
	// 	newItems = append(newItems, item)
	// }
	// page.Items = &newItems

	return page, nil
}

func (w *WithdrawalService) DisbursementTransaction(withdrawalID string, expenseID, cashBankID string) (err error) {
	if w.ctx.FinanceService == nil {
		return errors.New("finance service not found")
	}
	withdrawal := models.WithdrawalModel{}
	if err = w.db.Where("id = ?", withdrawalID).First(&withdrawal).Error; err != nil {
		return err
	}

	err = w.ctx.FinanceService.(*finance.FinanceService).TransactionService.CreateTransaction(&models.TransactionModel{
		SourceID:                    &expenseID,
		DestinationID:               &cashBankID,
		TransactionSecondaryRefType: "withdrawal",
		TransactionSecondaryRefID:   &withdrawal.ID,
		Description:                 "Withdrawal Disbursement - " + withdrawal.Code,
		Notes:                       withdrawal.Remarks,
	}, withdrawal.Total)

	return
}

func (w *WithdrawalService) CountWithdrawalByStatus(status string) (count int64, err error) {
	return count, w.db.Model(&models.WithdrawalModel{}).Where("status = ?", status).Count(&count).Error
}
