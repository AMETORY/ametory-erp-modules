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
			v.Pos.Status = "COMPLETE"
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
