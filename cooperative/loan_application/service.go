package loan_application

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/cooperative/cooperative_setting"
	"github.com/AMETORY/ametory-erp-modules/cooperative/saving"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type LoanApplicationService struct {
	db                        *gorm.DB
	ctx                       *context.ERPContext
	cooperativeSettingService *cooperative_setting.CooperativeSettingService
	savingService             *saving.SavingService
}

func NewLoanApplicationService(db *gorm.DB,
	ctx *context.ERPContext,
	cooperativeSettingService *cooperative_setting.CooperativeSettingService,
	savingService *saving.SavingService,
) *LoanApplicationService {
	return &LoanApplicationService{
		db:                        db,
		ctx:                       ctx,
		cooperativeSettingService: cooperativeSettingService,
		savingService:             savingService,
	}
}

func (l *LoanApplicationService) CreatePayment(input models.InstallmentPayment, loan models.LoanApplicationModel, userID *string) error {

	// Cek apakah status pinjaman sudah approved
	if loan.Status != "Disbursed" {
		return fmt.Errorf("loan must be disbursed before payment")
	}

	if input.PaymentAmount-input.TotalPaid < 0 {
		return errors.New("payment amount cannot be less than total paid")
	}
	var balance = input.PaymentAmount - input.TotalPaid
	if balance > 0 && balance < 1 {
		input.PaymentAmount = input.TotalPaid
		balance = 0
	}

	input.MemberID = loan.MemberID
	input.LoanApplicationID = &loan.ID
	err := l.db.Create(&input).Error
	if err != nil {
		return err
	}

	refID := utils.Uuid()
	// POKOK
	trans := &models.TransactionModel{
		CompanyID:            loan.CompanyID,
		UserID:               loan.UserID,
		Date:                 input.PaymentDate,
		LoanApplicationID:    &loan.ID,
		InstallmentPaymentID: &input.ID,
		CooperativeMemberID:  loan.MemberID,
		Debit:                input.PrincipalPaid,
		Credit:               0,
		Description:          fmt.Sprintf("Pembayaran Pokok Cicilan #%d [%s]", input.InstallmentNo, loan.LoanNumber),
		IsAccountReceivable:  true,
		IsLending:            true,
		AccountID:            loan.AccountReceivableID,
		TransactionRefID:     &refID,
		TransactionRefType:   "installment_payment",
	}
	if err := l.db.Create(&trans).Error; err != nil {
		return err
	}

	trans = &models.TransactionModel{

		CompanyID:            loan.CompanyID,
		UserID:               loan.UserID,
		Date:                 input.PaymentDate,
		LoanApplicationID:    &loan.ID,
		InstallmentPaymentID: &input.ID,
		CooperativeMemberID:  loan.MemberID,
		Credit:               input.PrincipalPaid,
		Debit:                0,
		Description:          fmt.Sprintf("Pembayaran Pokok Cicilan #%d [%s]", input.InstallmentNo, loan.LoanNumber),
		IsAccountReceivable:  false,
		IsLending:            true,
		AccountID:            loan.AccountAssetID,
		TransactionRefID:     &refID,
		TransactionRefType:   "installment_payment",
	}
	if err := l.db.Create(&trans).Error; err != nil {
		return err
	}

	// PROFIT
	if input.ProfitPaid > 0 {
		refID2 := utils.Uuid()
		trans = &models.TransactionModel{

			CompanyID:            loan.CompanyID,
			UserID:               loan.UserID,
			Date:                 input.PaymentDate,
			LoanApplicationID:    &loan.ID,
			InstallmentPaymentID: &input.ID,
			CooperativeMemberID:  loan.MemberID,
			Credit:               input.ProfitPaid,
			Debit:                0,
			Description:          fmt.Sprintf("Pembayaran Profit / Bunga Cicilan #%d [%s]", input.InstallmentNo, loan.LoanNumber),
			IsAccountReceivable:  false,
			IsLending:            true,
			AccountID:            loan.AccountIncomeID,
			TransactionRefID:     &refID2,
			TransactionRefType:   "installment_payment",
		}
		if err := l.db.Create(&trans).Error; err != nil {
			return err
		}

		trans = &models.TransactionModel{
			CompanyID:            loan.CompanyID,
			UserID:               loan.UserID,
			Date:                 input.PaymentDate,
			LoanApplicationID:    &loan.ID,
			InstallmentPaymentID: &input.ID,
			CooperativeMemberID:  loan.MemberID,
			Credit:               input.ProfitPaid,
			Debit:                0,
			Description:          fmt.Sprintf("Pembayaran Profit / Bunga Cicilan #%d [%s]", input.InstallmentNo, loan.LoanNumber),
			IsAccountReceivable:  false,
			IsLending:            true,
			AccountID:            loan.AccountAssetID,
			TransactionRefID:     &refID2,
			TransactionRefType:   "installment_payment",
		}
		if err := l.db.Create(&trans).Error; err != nil {
			return err
		}
	}
	// ADMIN
	if input.AdminFeePaid > 0 {
		refID3 := utils.Uuid()
		trans = &models.TransactionModel{

			CompanyID:            loan.CompanyID,
			UserID:               loan.UserID,
			Date:                 input.PaymentDate,
			LoanApplicationID:    &loan.ID,
			InstallmentPaymentID: &input.ID,
			CooperativeMemberID:  loan.MemberID,
			Credit:               input.AdminFeePaid,
			Debit:                0,
			Description:          fmt.Sprintf("Pembayaran Biaya Admin Cicilan #%d [%s]", input.InstallmentNo, loan.LoanNumber),
			IsAccountReceivable:  false,
			IsLending:            true,
			AccountID:            loan.AccountAdminFeeID,
			TransactionRefID:     &refID3,
			TransactionRefType:   "installment_payment",
		}
		if err := l.db.Create(&trans).Error; err != nil {
			return err
		}

		trans = &models.TransactionModel{

			CompanyID:            loan.CompanyID,
			UserID:               loan.UserID,
			Date:                 input.PaymentDate,
			LoanApplicationID:    &loan.ID,
			InstallmentPaymentID: &input.ID,
			CooperativeMemberID:  loan.MemberID,
			Credit:               input.AdminFeePaid,
			Debit:                0,
			Description:          fmt.Sprintf("Pembayaran Biaya Admin Cicilan #%d [%s]", input.InstallmentNo, loan.LoanNumber),
			IsAccountReceivable:  false,
			IsLending:            true,
			AccountID:            loan.AccountAssetID,
			TransactionRefID:     &refID3,
			TransactionRefType:   "installment_payment",
		}
		if err := l.db.Create(&trans).Error; err != nil {
			return err
		}
	}

	if balance > 0 {
		// CREATE Voluntry saving
		saving := models.SavingModel{

			CompanyID:            loan.CompanyID,
			UserID:               userID,
			MemberID:             loan.MemberID,
			AccountDestinationID: loan.AccountAssetID,
			SavingType:           "Voluntary",
			Amount:               balance,
			Notes:                fmt.Sprintf("Sisa Cicilan #%d [%s]", input.InstallmentNo, loan.LoanNumber),
			Date:                 &input.PaymentDate,
		}

		if err := l.db.Create(&saving).Error; err != nil {
			return err
		}

		err := l.savingService.CreateTransaction(saving, true)
		if err != nil {
			return err
		}
	}

	if input.RemainingLoan == 0 {
		loan.Status = "Settlement"
		if err := l.db.Save(&loan).Error; err != nil {
			return err
		}
	}
	return nil
}

func (l *LoanApplicationService) DisburseLoan(loan models.LoanApplicationModel, AccountAssetID *string, user *models.UserModel) error {

	// Cek apakah status pinjaman sudah approved
	if loan.Status != "Approved" {
		return fmt.Errorf("loan must be approved before disbursement")
	}
	if AccountAssetID == nil {
		return fmt.Errorf("account asset id is required")
	}
	if user == nil {
		return fmt.Errorf("user is required")
	}

	now := time.Now()

	// Set tanggal pencairan
	loan.DisbursementDate = &now
	loan.AccountAssetID = AccountAssetID
	loan.ApprovedBy = &user.FullName

	// Ubah status pinjaman menjadi "Disbursed"
	loan.Status = "Disbursed"
	loan.Remarks = loan.Remarks + "\n- [" + time.Now().Format("2006-01-02 15:04:05") + "] " + user.FullName + ": " + "Pencairan Pinjaman " + loan.LoanNumber

	if err := l.db.Create(&models.TransactionModel{
		CompanyID:           loan.CompanyID,
		UserID:              loan.UserID,
		Credit:              loan.LoanAmount,
		AccountID:           loan.AccountReceivableID,
		Description:         "Pencairan Pinjaman [" + loan.LoanNumber + "]",
		Date:                now,
		IsAccountReceivable: true,
		IsLending:           true,
		LoanApplicationID:   &loan.ID,
	}).Error; err != nil {
		return err
	}
	if err := l.db.Create(&models.TransactionModel{
		CompanyID:         loan.CompanyID,
		UserID:            loan.UserID,
		Debit:             loan.LoanAmount,
		AccountID:         loan.AccountAssetID,
		Description:       "Pencairan Pinjaman [" + loan.LoanNumber + "]",
		Date:              now,
		IsLending:         true,
		LoanApplicationID: &loan.ID,
	}).Error; err != nil {
		return err
	}

	return l.db.Save(&loan).Error
}

func (s *LoanApplicationService) GetLoans(request http.Request, search string, memberID *string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db
	if search != "" {
		stmt = stmt.Where("description ILIKE ? OR brands.name ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
		)
	}
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ? or company_id is null", request.Header.Get("ID-Company"))
	}
	if memberID != nil {
		stmt = stmt.Where("member_id = ?", memberID)
	}
	request.URL.Query().Get("page")
	stmt = stmt.Model(&models.LoanApplicationModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.LoanApplicationModel{})
	page.Page = page.Page + 1
	return page, nil
}

func (s *LoanApplicationService) GetLoanByID(id string, memberID *string) (*models.LoanApplicationModel, error) {
	var loan models.LoanApplicationModel
	db := s.db
	if memberID != nil {
		db = db.Where("member_id = ?", memberID)
	}
	if err := db.Where("id = ?", id).First(&loan).Error; err != nil {
		return nil, err
	}
	return &loan, nil
}
func (c *LoanApplicationService) CreateLoan(loan *models.LoanApplicationModel) error {
	return c.ctx.DB.Create(loan).Error
}

func (c *LoanApplicationService) UpdateLoan(id string, loan *models.LoanApplicationModel) error {
	return c.ctx.DB.Where("id = ?", id).Save(loan).Error
}

func (c *LoanApplicationService) DeleteLoan(id string) error {
	return c.ctx.DB.Delete(&models.LoanApplicationModel{}, "id = ?", id).Error
}
