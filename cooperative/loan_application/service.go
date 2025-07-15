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
	"github.com/AMETORY/ametory-erp-modules/finance"
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type LoanApplicationService struct {
	db                        *gorm.DB
	ctx                       *context.ERPContext
	cooperativeSettingService *cooperative_setting.CooperativeSettingService
	savingService             *saving.SavingService
	financeService            *finance.FinanceService
}

// NewLoanApplicationService creates a new instance of LoanApplicationService.
//
// The LoanApplicationService is used to manage loan applications, which are
// requests for loans made by members of a cooperative.
//
// It takes as parameters:
//   - db: a pointer to a GORM database instance
//   - ctx: a pointer to an ERPContext, which contains the user's HTTP request
//     context and other relevant information
//   - cooperativeSettingService: a pointer to a CooperativeSettingService,
//     which is used to look up the settings for the cooperative
//   - savingService: a pointer to a SavingService, which is used to manage
//     savings accounts
//   - financeService: a pointer to a FinanceService, which is used to manage
//     financial transactions
func NewLoanApplicationService(db *gorm.DB,
	ctx *context.ERPContext,
	cooperativeSettingService *cooperative_setting.CooperativeSettingService,
	savingService *saving.SavingService,
	financeService *finance.FinanceService,
) *LoanApplicationService {
	return &LoanApplicationService{
		db:                        db,
		ctx:                       ctx,
		cooperativeSettingService: cooperativeSettingService,
		savingService:             savingService,
		financeService:            financeService,
	}
}

// CreatePayment is used to create a new payment for a loan application.
//
// It takes as parameters:
//   - input: a pointer to an InstallmentPayment, which contains the details of the payment
//   - loan: a pointer to a LoanApplicationModel, which contains the details of the loan
//   - userID: a pointer to a string, which is the ID of the user making the payment
//
// It returns an error if the payment cannot be created.
func (l *LoanApplicationService) CreatePayment(input *models.InstallmentPayment, loan *models.LoanApplicationModel, userID *string) error {

	// Cek apakah status pinjaman sudah approved
	if loan.Status != "DISBURSED" {
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

	err := l.db.Transaction(func(tx *gorm.DB) error {
		l.financeService.TransactionService.SetDB(tx)
		l.savingService.SetDB(tx)

		input.MemberID = loan.MemberID
		input.LoanApplicationID = &loan.ID
		err := tx.Create(&input).Error
		if err != nil {

			return err
		}

		// refID := utils.Uuid()
		principalID := utils.Uuid()
		principalAssetID := utils.Uuid()
		// POKOK
		trans := &models.TransactionModel{
			Code:                        utils.RandString(10, false),
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
		if err := tx.Create(&trans).Error; err != nil {
			return err
		}

		accountAssetID := input.AccountAssetID
		if accountAssetID == nil {
			accountAssetID = loan.AccountAssetID
		}

		trans = &models.TransactionModel{
			Code:                        utils.RandString(10, false),
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
			AccountID:                   accountAssetID,
			TransactionRefID:            &principalID,
			TransactionRefType:          "transaction",
			TransactionSecondaryRefID:   &loan.ID,
			TransactionSecondaryRefType: "loan",
		}
		if err := tx.Create(&trans).Error; err != nil {
			return err
		}

		// PROFIT
		if input.ProfitPaid > 0 {
			profitID := utils.Uuid()
			profitAssetID := utils.Uuid()
			trans = &models.TransactionModel{
				Code:                        utils.RandString(10, false),
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
			if err := tx.Create(&trans).Error; err != nil {
				return err
			}

			trans = &models.TransactionModel{
				Code:                        utils.RandString(10, false),
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
				AccountID:                   accountAssetID,
				TransactionRefID:            &profitID,
				TransactionRefType:          "transaction",
				TransactionSecondaryRefID:   &loan.ID,
				TransactionSecondaryRefType: "loan",
			}
			if err := tx.Create(&trans).Error; err != nil {
				return err
			}
		}
		// ADMIN
		if input.AdminFeePaid > 0 {
			adminID := utils.Uuid()
			adminCashID := utils.Uuid()
			trans = &models.TransactionModel{
				Code:                        utils.RandString(10, false),
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
			if err := tx.Create(&trans).Error; err != nil {
				return err
			}

			trans = &models.TransactionModel{
				Code:                        utils.RandString(10, false),
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
				AccountID:                   accountAssetID,
				TransactionRefID:            &adminID,
				TransactionRefType:          "transaction",
				TransactionSecondaryRefID:   &loan.ID,
				TransactionSecondaryRefType: "loan",
			}
			if err := tx.Create(&trans).Error; err != nil {
				return err
			}
		}

		if balance > 0 {

			// CREATE Voluntry saving
			saving := models.SavingModel{
				CompanyID:            loan.CompanyID,
				UserID:               userID,
				CooperativeMemberID:  loan.MemberID,
				AccountDestinationID: accountAssetID,
				SavingType:           "VOLUNTARY",
				Amount:               balance,
				Notes:                fmt.Sprintf("Sisa Cicilan #%d [%s]", input.InstallmentNo, loan.LoanNumber),
				Date:                 &input.PaymentDate,
			}

			if err := tx.Create(&saving).Error; err != nil {
				return err
			}
			var company models.CompanyModel
			err := tx.Where("id = ?", loan.CompanyID).First(&company).Error
			if err != nil {
				return err
			}
			saving.Company = &company

			err = l.savingService.CreateTransaction(saving, true)
			if err != nil {
				return err
			}
		}

		if input.RemainingLoan == 0 {
			loan.Status = "SETTLEMENT"
			if err := tx.Save(&loan).Error; err != nil {
				return err
			}
		}

		return nil
	})

	l.financeService.TransactionService.SetDB(l.db)
	l.savingService.SetDB(l.db)
	if err != nil {
		return err
	}

	return nil
}

// DisburseLoan mencairkan pinjaman yang telah diapproved dan mengganti status pinjaman menjadi "DISBURSED".
// Fungsi ini juga membuat transaksi debet dan kredit yang terkait dengan pinjaman.
// Parameter accountAssetID harus diisi dengan ID account asset yang digunakan untuk mencairkan pinjaman.
// Parameter user harus diisi dengan user yang melakukan pencairan pinjaman.
// Parameter remarks harus diisi dengan keterangan pencairan pinjaman.
func (l *LoanApplicationService) DisburseLoan(loan *models.LoanApplicationModel, accountAssetID *string, user *models.UserModel, remarks string) error {
	if accountAssetID == nil {
		return errors.New("account asset id is required")
	}
	// Cek apakah status pinjaman sudah approved
	if loan.Status != "APPROVED" {
		return fmt.Errorf("loan must be approved before disbursement")
	}

	if user == nil {
		return fmt.Errorf("user is required")
	}

	now := time.Now()

	// Set tanggal pencairan
	loan.DisbursementDate = &now
	loan.AccountAssetID = accountAssetID
	loan.ApprovedBy = &user.FullName

	// Ubah status pinjaman menjadi "Disbursed"
	loan.Status = "DISBURSED"
	loan.Remarks = loan.Remarks + "[" + time.Now().Format("2006-01-02 15:04:05") + "] " + user.FullName + ": " + "Pencairan Pinjaman " + loan.LoanNumber + "\n" + remarks + "\n"
	transID := utils.Uuid()
	secTransID := utils.Uuid()
	return l.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&models.TransactionModel{
			Code:                        utils.RandString(10, false),
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
			CooperativeMemberID:         loan.MemberID,
		}).Error; err != nil {
			return err
		}
		if err := tx.Create(&models.TransactionModel{
			Code:                        utils.RandString(10, false),
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
			CooperativeMemberID:         loan.MemberID,
		}).Error; err != nil {
			return err
		}
		return tx.Where("id = ?", loan.ID).Updates(&loan).Error
	})

}

// GetLoans digunakan untuk mengambil data pinjaman yang ada di database yang sesuai dengan kriteria yang diberikan.
// Fungsi ini akan mengembalikan data pinjaman dalam bentuk paginate.Page.
//
// Parameter request digunakan untuk mengatur halaman dan per halaman berapa data yang akan diambil.
// Parameter search digunakan untuk mencari data pinjaman yang sesuai dengan kata kunci yang diberikan.
// Parameter memberID digunakan untuk mengambil data pinjaman yang sesuai dengan ID anggota yang diberikan.
//
// Fungsi ini akan mengembalikan error jika terjadi kesalahan saat mengambil data dari database.
func (s *LoanApplicationService) GetLoans(request http.Request, search string, memberID *string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.Preload("Member")
	if search != "" {
		stmt = stmt.Where("remarks ILIKE ?",
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

// GetLoanByID mengambil data pinjaman berdasarkan ID yang diberikan.
// Jika parameter memberID diisi maka akan mengambil data pinjaman yang sesuai dengan ID anggota yang diberikan.
//
// Fungsi ini akan mengembalikan error jika terjadi kesalahan saat mengambil data dari database.
func (s *LoanApplicationService) GetLoanByID(id string, memberID *string) (*models.LoanApplicationModel, error) {
	var loan models.LoanApplicationModel
	db := s.db.Preload(clause.Associations)
	if memberID != nil {
		db = db.Where("member_id = ?", memberID)
	}
	if err := db.Where("id = ?", id).First(&loan).Error; err != nil {
		return nil, err
	}
	preview := s.GetPreview(&loan)
	loan.Preview = preview
	err := s.GetTransactions(&loan)
	if err != nil {
		return nil, err
	}
	if loan.Data != "" {
		err := json.Unmarshal([]byte(loan.Data), &loan.Installments)
		if err != nil {
			return nil, err
		}
	}
	return &loan, err
}

// CreateLoan digunakan untuk membuat pinjaman baru. Fungsi ini akan mengembalikan error jika terjadi kesalahan saat membuat pinjaman.
// Fungsi ini akan mengatur tipe pinjaman dan tingkat bunga/profit rate pinjaman berdasarkan setting perusahaan yang diberikan.
// Fungsi ini juga akan mengatur nomor pinjaman berdasarkan setting perusahaan yang diberikan.
func (c *LoanApplicationService) CreateLoan(loan *models.LoanApplicationModel) error {
	loan.Data = "[]"
	loan.Status = "DRAFT"
	setting, err := c.cooperativeSettingService.GetSetting(loan.CompanyID)
	if err != nil {
		return err
	}
	if setting.IsIslamic {
		loan.LoanType = "MUDHARABAH"
		loan.ExpectedProfitRate = float64(loan.RepaymentTerm) * setting.ExpectedProfitRatePerMonth
	} else {
		loan.LoanType = "CONVENTIONAL"
		loan.ProfitType = "ANUITY"
		loan.InterestRate = float64(loan.RepaymentTerm) * setting.InterestRatePerMonth
	}

	loan.AccountAdminFeeID = setting.LoanAccountAdminFeeID
	loan.AccountIncomeID = setting.LoanAccountIncomeID
	loan.AccountReceivableID = setting.LoanAccountID
	loan.TermCondition = setting.TermCondition
	err = c.GenNumber(loan, loan.CompanyID)
	if err != nil {
		return err
	}

	return c.db.Create(loan).Error
}

// UpdateLoan digunakan untuk mengupdate pinjaman yang sudah ada. Fungsi ini akan mengembalikan error jika terjadi kesalahan saat mengupdate pinjaman.
// Fungsi ini akan mengatur tipe pinjaman dan tingkat bunga/profit rate pinjaman berdasarkan setting perusahaan yang diberikan.
// Fungsi ini juga akan mengatur nomor pinjaman berdasarkan setting perusahaan yang diberikan.
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

// DeleteLoan menghapus pinjaman yang sesuai dengan ID yang diberikan.
// Fungsi ini akan mengembalikan error jika terjadi kesalahan saat menghapus pinjaman.
// Fungsi ini juga akan menghapus transaksi yang terkait dengan pinjaman yang dihapus.
func (s *LoanApplicationService) DeleteLoan(id string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Where("loan_application_id = ?", id).Delete(&models.TransactionModel{}).Error
		if err != nil {
			return err
		}
		return tx.Where("id = ?", id).Delete(&models.LoanApplicationModel{}).Error
	})
}

// ApprovalLoan digunakan untuk menyetujui atau menolak pinjaman yang sesuai dengan ID yang diberikan.
// Fungsi ini akan mengembalikan error jika terjadi kesalahan saat menyetujui atau menolak pinjaman.
// Fungsi ini juga akan mengatur status pinjaman menjadi "APPROVED" atau "REJECTED".
// Jika status yang diberikan adalah "APPROVED" maka akan mengenerate tabel cicilan dan mengisi kolom "data" di pinjaman dengan hasil generate tabel cicilan.
// Jika status yang diberikan adalah "REJECTED" maka akan mengisi kolom "remarks" di pinjaman dengan keterangan penolakan pinjaman.
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
		loan.Remarks = loan.Remarks + "[" + time.Now().Format("2006-01-02 15:04:05") + "] " + user.FullName + ": " + "Persetujuan " + loan.LoanNumber + "\n" + remarks + "\n"
		loan.Data = string(b)
	}

	if status == "REJECTED" {
		loan.Remarks = loan.Remarks + "[" + time.Now().Format("2006-01-02 15:04:05") + "] " + user.FullName + ": " + "Penolakan " + loan.LoanNumber + "\n" + remarks + "\n"
	}

	loan.Status = status
	return c.db.Where("id =?", id).Save(loan).Error
}

// GenerateInstallmentTable generates installment table based on loan application model.
// It supports MUDHARABAH, QARDH_HASAN, and CONVENTIONAL loan types.
// CONVENTIONAL loan type supports ANUITY, FIXED, and DECLINING profit type.
// It returns an array of models.InstallmentDetail and error.
// The installment detail contains installment number, principal amount, interest amount, admin fee, total paid, and remaining loan.
// The function will calculate the installment table based on the loan application model and return the result.
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
		case "MUDHARABAH":
			// interestAmount = remainingLoan * loan.ExpectedProfitRate / 100 / float64(loan.RepaymentTerm)
			// principalAmount = loan.LoanAmount / float64(loan.RepaymentTerm)
			// totalPaid = principalAmount + interestAmount + fixedAdminFee
			interestAmount = totalProfit / float64(loan.RepaymentTerm)
			principalAmount = loan.LoanAmount / float64(loan.RepaymentTerm)
			totalPaid = principalAmount + interestAmount + fixedAdminFee

		case "QARDH_HASAN":
			principalAmount = loan.LoanAmount / float64(loan.RepaymentTerm)
			totalPaid = principalAmount + fixedAdminFee
		default:
			switch loan.ProfitType {
			case "ANUITY":
				if loan.InterestRate > 0 {
					interestAmount = remainingLoan * interestRateMonthly
					principalAmount = annuity - interestAmount
					totalPaid = principalAmount + interestAmount + fixedAdminFee
				} else {
					principalAmount = loan.LoanAmount / float64(loan.RepaymentTerm)
					totalPaid = principalAmount + fixedAdminFee
				}

			case "FIXED":
				interestAmount = loan.LoanAmount * loan.InterestRate / 100 / float64(loan.RepaymentTerm)
				principalAmount = loan.LoanAmount / float64(loan.RepaymentTerm)
				totalPaid = principalAmount + interestAmount + fixedAdminFee
			case "DECLINING":
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

// GetPreview returns a map of string to []InstallmentDetail containing the preview of
// the installment table for each type of loan (ANUITY, DECLINING, FIXED, QARDH_HASAN, MUDHARABAH).
// The map key is the type of loan, and the value is the preview of the installment table.
func (c *LoanApplicationService) GetPreview(loan *models.LoanApplicationModel) map[string][]models.InstallmentDetail {

	if loan.LoanType == "CONVENTIONAL" {
		loan.ProfitType = "ANUITY"
		anuityTable, err := c.GenerateInstallmentTable(loan)
		if err != nil {
			fmt.Println(err)
		}
		loan.ProfitType = "DECLINING"
		decliningTable, err := c.GenerateInstallmentTable(loan)
		if err != nil {
			fmt.Println(err)
		}
		loan.ProfitType = "FIXED"
		fixedTable, err := c.GenerateInstallmentTable(loan)
		if err != nil {
			fmt.Println(err)
		}
		// fmt.Println(map[string][]InstallmentDetail{
		// 	"anuity":    anuityTable,
		// 	"declining": decliningTable,
		// })
		return map[string][]models.InstallmentDetail{
			"ANUITY":    anuityTable,
			"DECLINING": decliningTable,
			"FIXED":     fixedTable,
		}
	}

	loan.LoanType = "QARDH_HASAN"
	qardhHasanTable, _ := c.GenerateInstallmentTable(loan)
	loan.LoanType = "MUDHARABAH"
	mudharabahTable, _ := c.GenerateInstallmentTable(loan)

	return map[string][]models.InstallmentDetail{
		"QARDH_HASAN": qardhHasanTable,
		"MUDHARABAH":  mudharabahTable,
	}
}

// GenNumber generates the next number for a new loan. It queries the database to get the latest
// loan number for the given company, and then uses the invoice bill setting to generate the next
// number. If the query fails, it falls back to generating the number from the invoice bill setting
// with a prefix of "00".
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

// GetMember retrieves a cooperative member by its ID and stores it in the
// LoanApplicationModel under the Member field. It does not return an error.
func (c *LoanApplicationService) GetMember(loan *models.LoanApplicationModel) {
	member := models.CooperativeMemberModel{}
	c.db.First(&member, "id = ?", loan.MemberID)
	loan.Member = &member
}

// GetTransactions retrieves the transactions and payments for a loan
// and stores them in the LoanApplicationModel under the Transactions and
// Payments fields respectively. If the loan has payments, it also retrieves
// the last payment and stores it in the LoanApplicationModel under the
// LastPayment field. It returns an error if the query fails.
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
	if err == nil {
		loan.LastPayment = &lastPayment
	}

	return nil
}
