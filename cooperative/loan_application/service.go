package loan_application

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/cooperative/cooperative_setting"
	"github.com/AMETORY/ametory-erp-modules/cooperative/saving"
	"github.com/AMETORY/ametory-erp-modules/shared"
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

func (l *LoanApplicationService) CreatePayment(input *models.InstallmentPayment, loan *models.LoanApplicationModel, userID *string) error {

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

	// refID := utils.Uuid()
	principalID := utils.Uuid()
	principalAssetID := utils.Uuid()
	// POKOK
	trans := &models.TransactionModel{
		BaseModel:                   shared.BaseModel{ID: principalID},
		CompanyID:                   loan.CompanyID,
		UserID:                      loan.UserID,
		Date:                        input.PaymentDate,
		LoanApplicationID:           &loan.ID,
		InstallmentPaymentID:        &input.ID,
		CooperativeMemberID:         loan.MemberID,
		Credit:                      input.PrincipalPaid,
		Description:                 fmt.Sprintf("Pembayaran Pokok Cicilan #%d [%s]", input.InstallmentNo, loan.LoanNumber),
		IsAccountReceivable:         true,
		IsLending:                   true,
		AccountID:                   loan.AccountReceivableID,
		TransactionRefID:            &principalAssetID,
		TransactionRefType:          "transaction",
		TransactionSecondaryRefID:   &loan.ID,
		TransactionSecondaryRefType: "loan",
	}
	if err := l.db.Create(&trans).Error; err != nil {
		return err
	}

	trans = &models.TransactionModel{
		BaseModel:                   shared.BaseModel{ID: principalAssetID},
		CompanyID:                   loan.CompanyID,
		UserID:                      loan.UserID,
		Date:                        input.PaymentDate,
		LoanApplicationID:           &loan.ID,
		InstallmentPaymentID:        &input.ID,
		CooperativeMemberID:         loan.MemberID,
		Debit:                       input.PrincipalPaid,
		Description:                 fmt.Sprintf("Pembayaran Pokok Cicilan #%d [%s]", input.InstallmentNo, loan.LoanNumber),
		IsAccountReceivable:         false,
		IsLending:                   true,
		AccountID:                   loan.AccountAssetID,
		TransactionRefID:            &principalID,
		TransactionRefType:          "transaction",
		TransactionSecondaryRefID:   &loan.ID,
		TransactionSecondaryRefType: "loan",
	}
	if err := l.db.Create(&trans).Error; err != nil {
		return err
	}

	// PROFIT
	if input.ProfitPaid > 0 {
		profitID := utils.Uuid()
		profitAssetID := utils.Uuid()
		trans = &models.TransactionModel{
			BaseModel:                   shared.BaseModel{ID: profitID},
			CompanyID:                   loan.CompanyID,
			UserID:                      loan.UserID,
			Date:                        input.PaymentDate,
			LoanApplicationID:           &loan.ID,
			InstallmentPaymentID:        &input.ID,
			CooperativeMemberID:         loan.MemberID,
			Credit:                      input.ProfitPaid,
			Description:                 fmt.Sprintf("Pembayaran Profit / Bunga Cicilan #%d [%s]", input.InstallmentNo, loan.LoanNumber),
			IsAccountReceivable:         false,
			IsLending:                   true,
			AccountID:                   loan.AccountIncomeID,
			TransactionRefID:            &profitAssetID,
			TransactionRefType:          "transaction",
			TransactionSecondaryRefID:   &loan.ID,
			TransactionSecondaryRefType: "loan",
		}
		if err := l.db.Create(&trans).Error; err != nil {
			return err
		}

		trans = &models.TransactionModel{
			BaseModel:                   shared.BaseModel{ID: profitAssetID},
			CompanyID:                   loan.CompanyID,
			UserID:                      loan.UserID,
			Date:                        input.PaymentDate,
			LoanApplicationID:           &loan.ID,
			InstallmentPaymentID:        &input.ID,
			CooperativeMemberID:         loan.MemberID,
			Debit:                       input.ProfitPaid,
			Description:                 fmt.Sprintf("Pembayaran Profit / Bunga Cicilan #%d [%s]", input.InstallmentNo, loan.LoanNumber),
			IsAccountReceivable:         false,
			IsLending:                   true,
			AccountID:                   loan.AccountAssetID,
			TransactionRefID:            &profitID,
			TransactionRefType:          "transaction",
			TransactionSecondaryRefID:   &loan.ID,
			TransactionSecondaryRefType: "loan",
		}
		if err := l.db.Create(&trans).Error; err != nil {
			return err
		}
	}
	// ADMIN
	if input.AdminFeePaid > 0 {
		adminID := utils.Uuid()
		adminCashID := utils.Uuid()
		trans = &models.TransactionModel{
			BaseModel:                   shared.BaseModel{ID: adminID},
			CompanyID:                   loan.CompanyID,
			UserID:                      loan.UserID,
			Date:                        input.PaymentDate,
			LoanApplicationID:           &loan.ID,
			InstallmentPaymentID:        &input.ID,
			CooperativeMemberID:         loan.MemberID,
			Credit:                      input.AdminFeePaid,
			Description:                 fmt.Sprintf("Pembayaran Biaya Admin Cicilan #%d [%s]", input.InstallmentNo, loan.LoanNumber),
			IsAccountReceivable:         false,
			IsLending:                   true,
			AccountID:                   loan.AccountAdminFeeID,
			TransactionRefID:            &adminCashID,
			TransactionRefType:          "transaction",
			TransactionSecondaryRefID:   &loan.ID,
			TransactionSecondaryRefType: "loan",
		}
		if err := l.db.Create(&trans).Error; err != nil {
			return err
		}

		trans = &models.TransactionModel{
			BaseModel:                   shared.BaseModel{ID: adminCashID},
			CompanyID:                   loan.CompanyID,
			UserID:                      loan.UserID,
			Date:                        input.PaymentDate,
			LoanApplicationID:           &loan.ID,
			InstallmentPaymentID:        &input.ID,
			CooperativeMemberID:         loan.MemberID,
			Debit:                       input.AdminFeePaid,
			Description:                 fmt.Sprintf("Pembayaran Biaya Admin Cicilan #%d [%s]", input.InstallmentNo, loan.LoanNumber),
			IsAccountReceivable:         false,
			IsLending:                   true,
			AccountID:                   loan.AccountAssetID,
			TransactionRefID:            &adminID,
			TransactionRefType:          "transaction",
			TransactionSecondaryRefID:   &loan.ID,
			TransactionSecondaryRefType: "loan",
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
			SavingType:           "VOLUNTARY",
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

func (l *LoanApplicationService) DisburseLoan(loan *models.LoanApplicationModel, AccountAssetID *string, user *models.UserModel) error {

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
	transID := utils.Uuid()
	secTransID := utils.Uuid()

	if err := l.db.Create(&models.TransactionModel{
		BaseModel:                   shared.BaseModel{ID: transID},
		CompanyID:                   loan.CompanyID,
		UserID:                      loan.UserID,
		Debit:                       loan.LoanAmount,
		AccountID:                   loan.AccountReceivableID,
		Description:                 "Pencairan Pinjaman [" + loan.LoanNumber + "]",
		Date:                        now,
		IsAccountReceivable:         true,
		IsLending:                   true,
		LoanApplicationID:           &loan.ID,
		TransactionRefID:            &secTransID,
		TransactionRefType:          "transaction",
		TransactionSecondaryRefID:   &loan.ID,
		TransactionSecondaryRefType: "loan",
	}).Error; err != nil {
		return err
	}
	if err := l.db.Create(&models.TransactionModel{
		BaseModel:                   shared.BaseModel{ID: secTransID},
		CompanyID:                   loan.CompanyID,
		UserID:                      loan.UserID,
		Credit:                      loan.LoanAmount,
		AccountID:                   loan.AccountAssetID,
		Description:                 "Pencairan Pinjaman [" + loan.LoanNumber + "]",
		Date:                        now,
		IsLending:                   true,
		LoanApplicationID:           &loan.ID,
		TransactionRefID:            &transID,
		TransactionRefType:          "transaction",
		TransactionSecondaryRefID:   &loan.ID,
		TransactionSecondaryRefType: "loan",
	}).Error; err != nil {
		return err
	}

	return l.db.Save(&loan).Error
}

func (s *LoanApplicationService) GetLoans(request http.Request, search string, memberID *string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db
	if search != "" {
		stmt = stmt.Where("description ILIKE ?",
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
	preview := s.GetPreview(&loan)
	loan.Preview = preview
	err := s.GetTransactions(&loan)
	return &loan, err
}
func (c *LoanApplicationService) CreateLoan(loan *models.LoanApplicationModel) error {
	return c.db.Create(loan).Error
}

func (c *LoanApplicationService) UpdateLoan(id string, loan *models.LoanApplicationModel) error {

	err := c.db.Where("id = ?", id).Save(loan).Error
	if err != nil {
		return err
	}
	if loan.AdminFee == 0 {
		if err := c.db.Model(loan).Where("id =?", id).Update("admin_fee", 0).Error; err != nil {
			return err
		}
	}
	if loan.InterestRate == 0 {
		if err := c.db.Model(loan).Where("id =?", id).Update("interest_rate", 0).Error; err != nil {
			return err
		}
	}
	if loan.ExpectedProfitRate == 0 {
		if err := c.db.Model(loan).Where("id =?", id).Update("expected_profit_rate", 0).Error; err != nil {
			return err
		}
	}
	return nil
}

func (c *LoanApplicationService) DeleteLoan(id string) error {
	return c.db.Delete(&models.LoanApplicationModel{}, "id = ?", id).Error
}

func (c *LoanApplicationService) ApprovalLoan(id, userID, status, remarks string) error {

	var user models.UserModel
	if err := c.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return err
	}
	loan, err := c.GetLoanByID(id, nil)
	if err != nil {
		return err
	}
	if loan.AccountReceivableID == nil {
		return errors.New("account receivable id is empty")
	}
	if loan.AccountIncomeID == nil {
		return errors.New("account income id is empty")
	}
	if loan.AccountAsset == nil {
		// return errors.New("account asset id is empty")
	}
	if loan.AccountAdminFeeID == nil {
		return errors.New("account asset id is empty")
	}
	if status == "APPROVED" {
		installments, err := c.GenerateInstallmentTable(loan)
		if err != nil {
			return err

		}
		// fmt.Println(installments)
		b, err := json.Marshal(installments)
		if err != nil {
			return err
		}
		loan.Remarks = loan.Remarks + "\n- [" + time.Now().Format("2006-01-02 15:04:05") + "] " + user.FullName + ": " + "Persetujuan " + loan.LoanNumber
		loan.Data = string(b)
	}

	if status == "REJECTED" {
		loan.Remarks = loan.Remarks + "\n- [" + time.Now().Format("2006-01-02 15:04:05") + "] " + user.FullName + ": " + "Penolakan " + loan.LoanNumber + "\n" + remarks
	}

	return c.db.Where("id =?", id).Save(loan).Error
}
func (c *LoanApplicationService) GenerateInstallmentTable(loan *models.LoanApplicationModel) ([]models.InstallmentDetail, error) {
	table := []models.InstallmentDetail{}
	remainingLoan := loan.LoanAmount
	fixedAdminFee := loan.AdminFee

	// calculate annuity
	interestRateMonthly := loan.InterestRate / 100 / float64(loan.RepaymentTerm)
	annuityFactor := (math.Pow(1+interestRateMonthly, float64(loan.RepaymentTerm)) * interestRateMonthly) / (math.Pow(1+interestRateMonthly, float64(loan.RepaymentTerm)) - 1)
	annuity := loan.LoanAmount * annuityFactor

	totalProfit := loan.ProjectedProfit * loan.ExpectedProfitRate / 100
	// totalDebt := loan.LoanAmount + totalProfit
	// monthlyPayment := totalDebt / float64(loan.RepaymentTerm)

	// remainingDebt := totalDebt

	for i := 1; i <= loan.RepaymentTerm; i++ {
		var interestAmount, totalPaid, principalAmount float64

		switch loan.LoanType {
		case "Mudharabah":
			// interestAmount = remainingLoan * loan.ExpectedProfitRate / 100 / float64(loan.RepaymentTerm)
			// principalAmount = loan.LoanAmount / float64(loan.RepaymentTerm)
			// totalPaid = principalAmount + interestAmount + fixedAdminFee
			interestAmount = totalProfit / float64(loan.RepaymentTerm)
			principalAmount = loan.LoanAmount / float64(loan.RepaymentTerm)
			totalPaid = principalAmount + interestAmount + fixedAdminFee

		case "Qardh Hasan":
			principalAmount = loan.LoanAmount / float64(loan.RepaymentTerm)
			totalPaid = principalAmount + fixedAdminFee
		default:
			switch loan.ProfitType {
			case "anuity":
				if loan.InterestRate > 0 {
					interestAmount = remainingLoan * interestRateMonthly
					principalAmount = annuity - interestAmount
					totalPaid = principalAmount + interestAmount + fixedAdminFee
				} else {
					principalAmount = loan.LoanAmount / float64(loan.RepaymentTerm)
					totalPaid = principalAmount + fixedAdminFee
				}

			case "fixed":
				interestAmount = loan.LoanAmount * loan.InterestRate / 100 / float64(loan.RepaymentTerm)
				principalAmount = loan.LoanAmount / float64(loan.RepaymentTerm)
				totalPaid = principalAmount + interestAmount + fixedAdminFee
			case "declining":
				interestAmount = remainingLoan * loan.InterestRate / 100 / float64(loan.RepaymentTerm)
				principalAmount = loan.LoanAmount / float64(loan.RepaymentTerm)
				totalPaid = principalAmount + interestAmount + fixedAdminFee
			default:
				return nil, fmt.Errorf("unsupported profit type: %s", loan.ProfitType)
			}
		}

		if i == loan.RepaymentTerm {
			remainingLoan = 0
		} else {
			remainingLoan -= principalAmount
		}

		table = append(table, models.InstallmentDetail{
			InstallmentNumber: i,
			PrincipalAmount:   utils.AmountRound(principalAmount, 2),
			InterestAmount:    utils.AmountRound(interestAmount, 2),
			AdminFee:          utils.AmountRound(fixedAdminFee, 2),
			TotalPaid:         utils.AmountRound(totalPaid, 2),
			RemainingLoan:     utils.AmountRound(remainingLoan, 2),
		})

	}

	return table, nil
}

func (c *LoanApplicationService) GetPreview(loan *models.LoanApplicationModel) map[string][]models.InstallmentDetail {

	if loan.LoanType == "Conventional" {
		loan.ProfitType = "anuity"
		anuityTable, err := c.GenerateInstallmentTable(loan)
		if err != nil {
			fmt.Println(err)
		}
		loan.ProfitType = "declining"
		decliningTable, err := c.GenerateInstallmentTable(loan)
		if err != nil {
			fmt.Println(err)
		}
		loan.ProfitType = "fixed"
		fixedTable, err := c.GenerateInstallmentTable(loan)
		if err != nil {
			fmt.Println(err)
		}
		// fmt.Println(map[string][]InstallmentDetail{
		// 	"anuity":    anuityTable,
		// 	"declining": decliningTable,
		// })
		return map[string][]models.InstallmentDetail{
			"anuity":    anuityTable,
			"declining": decliningTable,
			"fixed":     fixedTable,
		}
	}

	loan.LoanType = "Qardh Hasan"
	qardhHasanTable, _ := c.GenerateInstallmentTable(loan)
	loan.LoanType = "Mudharabah"
	mudharabahTable, _ := c.GenerateInstallmentTable(loan)

	return map[string][]models.InstallmentDetail{
		"Qardh Hasan": qardhHasanTable,
		"Mudharabah":  mudharabahTable,
	}
}

func (c *LoanApplicationService) GenNumber(loan *models.LoanApplicationModel, companyID *string) error {
	setting, err := c.cooperativeSettingService.GetSetting(companyID)
	if err != nil {
		return err
	}
	lastLoan := models.LoanApplicationModel{}
	nextNumber := ""
	data := shared.InvoiceBillSettingModel{
		StaticCharacter:       setting.StaticCharacter,
		NumberFormat:          setting.NumberFormat,
		AutoNumericLength:     setting.AutoNumericLength,
		RandomNumericLength:   setting.RandomNumericLength,
		RandomCharacterLength: setting.RandomCharacterLength,
	}
	if err := c.db.Where("company_id = ?", companyID).Limit(1).Order("created_at desc").Find(&lastLoan).Error; err != nil {
		nextNumber = shared.GenerateInvoiceBillNumber(data, "00")
	} else {
		nextNumber = shared.ExtractNumber(data, lastLoan.LoanNumber)
	}

	loan.LoanNumber = nextNumber
	return nil
}
func (c *LoanApplicationService) GetMember(loan *models.LoanApplicationModel) {
	member := models.CooperativeMemberModel{}
	c.db.First(&member, "id = ?", loan.MemberID)
	loan.Member = &member
}
func (c *LoanApplicationService) GetTransactions(loan *models.LoanApplicationModel) error {
	transactions := []models.TransactionModel{}
	err := c.db.Find(&transactions, "loan_application_id = ?", loan.ID).Error
	if err != nil {
		return err
	}
	loan.Transactions = transactions
	var payments []models.InstallmentPayment
	err = c.db.Where("loan_application_id = ?", loan.ID).Find(&payments).Error
	if err != nil {
		return err
	}
	loan.Payments = payments

	var lastPayment models.InstallmentPayment
	err = c.db.Where("loan_application_id = ?", loan.ID).Order("installment_no DESC").First(&lastPayment).Error
	if err != nil {
		return err
	}

	loan.LastPayment = &lastPayment
	return err
}
