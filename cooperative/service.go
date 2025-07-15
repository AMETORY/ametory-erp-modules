package cooperative

import (
	"github.com/AMETORY/ametory-erp-modules/company"
	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/cooperative/cooperative_member"
	"github.com/AMETORY/ametory-erp-modules/cooperative/cooperative_setting"
	"github.com/AMETORY/ametory-erp-modules/cooperative/loan_application"
	"github.com/AMETORY/ametory-erp-modules/cooperative/net_surplus"
	"github.com/AMETORY/ametory-erp-modules/cooperative/saving"
	"github.com/AMETORY/ametory-erp-modules/finance"
	"github.com/AMETORY/ametory-erp-modules/finance/account"
	"github.com/AMETORY/ametory-erp-modules/shared/constants"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
)

type CooperativeService struct {
	ctx                       *context.ERPContext
	CompanyService            *company.CompanyService
	CooperativeSettingService *cooperative_setting.CooperativeSettingService
	CooperativeMemberService  *cooperative_member.CooperativeMemberService
	LoanApplicationService    *loan_application.LoanApplicationService
	SavingService             *saving.SavingService
	NetSurplusService         *net_surplus.NetSurplusService
	FinanceService            *finance.FinanceService
}

// NewCooperativeService creates a new instance of CooperativeService.
//
// The service provides methods for managing the cooperative feature of the
// application. It requires a pointer to a GORM database instance, a pointer to
// an ERPContext, a pointer to a CompanyService, and a pointer to a FinanceService.
//
// The service instance is then used to initialize other services, such as
// CooperativeSettingService, CooperativeMemberService, LoanApplicationService,
// SavingService, and NetSurplusService.
func NewCooperativeService(ctx *context.ERPContext, companySrv *company.CompanyService, financeService *finance.FinanceService) *CooperativeService {
	cooperativeSettingService := cooperative_setting.NewCooperativeSettingService(ctx.DB, ctx)
	savingService := saving.NewSavingService(ctx.DB, ctx, cooperativeSettingService)
	var service = CooperativeService{
		ctx:                       ctx,
		CompanyService:            companySrv,
		CooperativeSettingService: cooperativeSettingService,
		CooperativeMemberService:  cooperative_member.NewCooperativeMemberService(ctx.DB, ctx),
		SavingService:             savingService,
		LoanApplicationService:    loan_application.NewLoanApplicationService(ctx.DB, ctx, cooperativeSettingService, savingService, financeService),
		NetSurplusService:         net_surplus.NewNetSurplusService(ctx.DB, ctx, cooperativeSettingService, financeService, savingService),
		FinanceService:            financeService,
	}
	if err := service.Migrate(); err != nil {
		panic(err)
	}
	return &service
}

// Migrate migrates the database schema for the CooperativeService.
//
// If the SkipMigration flag is set to true in the context, this method
// will not perform any migration and will return nil. Otherwise, it will
// attempt to auto-migrate the database to include the schemas for
// CooperativeSettingModel, CooperativeMemberModel, LoanApplicationModel,
// SavingModel, InstallmentPayment, NetSurplusModel, and MemberInvitationModel.
// If the migration process encounters an error, it will return that error.
// Otherwise, it will return nil upon successful migration.
func (s *CooperativeService) Migrate() error {
	if s.ctx.SkipMigration {
		return nil
	}
	return s.ctx.DB.AutoMigrate(
		&models.CooperativeSettingModel{},
		&models.CooperativeMemberModel{},
		&models.LoanApplicationModel{},
		&models.SavingModel{},
		&models.InstallmentPayment{},
		&models.NetSurplusModel{},
		&models.MemberInvitationModel{},
	)
}

// IslamicCooperationAccountsTemplate retrieves a list of Islamic cooperation account models.
//
// It assigns the provided userID and companyID to each account model in the list
// obtained from the predefined IslamicCooperationAccountsTemplate. Additionally, it
// appends accounts related to net surplus to the list, using the provided userID
// and companyID for these accounts as well.
//
// Parameters:
//   - userID: a pointer to the user ID associated with the account models.
//   - companyID: a pointer to the company ID associated with the account models.
//
// Returns a slice of AccountModel populated with the userID and companyID.

func (c *CooperativeService) IslamicCooperationAccountsTemplate(userID, companyID *string) []models.AccountModel {
	accounts := account.IslamicCooperationAccountsTemplate
	for i := range accounts {
		accounts[i].UserID = userID
		accounts[i].CompanyID = companyID
	}

	accounts = append(accounts, c.netSurplusAccounts(userID, companyID)...)

	// for i := range accounts {
	// 	accounts[i].ConvertTerm()
	// }

	return accounts
}

// CooperationAccountsTemplate retrieves a list of cooperation account models.
//
// It assigns the provided userID and companyID to each account model in the list
// obtained from the predefined CooperationAccountsTemplate. Additionally, it
// appends accounts related to net surplus to the list, using the provided userID
// and companyID for these accounts as well.
//
// Parameters:
//   - userID: a pointer to the user ID associated with the account models.
//   - companyID: a pointer to the company ID associated with the account models.
//
// Returns a slice of AccountModel populated with the userID and companyID.

func (c *CooperativeService) CooperationAccountsTemplate(userID, companyID *string) []models.AccountModel {
	accounts := account.CooperationAccountsTemplate
	for i := range accounts {
		accounts[i].UserID = userID
		accounts[i].CompanyID = companyID
	}
	accounts = append(accounts, c.netSurplusAccounts(userID, companyID)...)

	// for i := range accounts {
	// 	accounts[i].ConvertTerm()
	// }

	return accounts
}

// netSurplusAccounts returns a list of AccountModel that are related to net surplus.
//
// It returns the following accounts:
//
// - Cadangan Koperasi (33001)
// - Dana Pendidikan (33002)
// - Dana Sosial (33003)
// - Jasa Usaha (33004)
// - Jasa Modal (33005)
// - Dana Pengurus (33006)
// - Dana Lainnya (33007)
// - Hutang SHU Ke Anggota (23004)
// - Beban Sosial (53001)
// - Beban Pendidikan (53002)
// - Kas Khusus Alokasi SHU (13001)
// - Kas Alokasi SHU Jasa Usaha (13002)
// - Kas Alokasi SHU Jasa Modal (13003)
// - Kas Alokasi SHU Dana Sosial (13004)
// - Kas Alokasi SHU Dana Pendidikan (13005)
// - Kas Alokasi SHU Dana Cadangan (13006)
// - Kas Alokasi SHU Dana Pengurus (13007)
func (c *CooperativeService) netSurplusAccounts(userID, companyID *string) []models.AccountModel {
	return []models.AccountModel{
		{
			CashflowSubGroup: constants.OTHER_INVESTMENT_ACTIVITIES,
			CashflowGroup:    constants.CASHFLOW_GROUP_INVESTING,
			Category:         constants.CATEGORY_EQUITY,

			Name:      "Cadangan Koperasi",
			Type:      constants.TYPE_EQUITY,
			UserID:    userID,
			CompanyID: companyID,
			Code:      "33001",
		},
		{
			CashflowSubGroup: constants.OPERATIONAL_EXPENSES,
			CashflowGroup:    constants.CASHFLOW_GROUP_OPERATING,
			Category:         constants.CATEGORY_EQUITY,

			Name:      "Dana Pendidikan",
			Type:      constants.TYPE_EQUITY,
			UserID:    userID,
			CompanyID: companyID,
			Code:      "33002",
		},
		{
			CashflowSubGroup: constants.OPERATIONAL_EXPENSES,
			CashflowGroup:    constants.CASHFLOW_GROUP_OPERATING,
			Category:         constants.CATEGORY_EQUITY,

			Name:      "Dana Sosial",
			Type:      constants.TYPE_EQUITY,
			UserID:    userID,
			CompanyID: companyID,
			Code:      "33003",
		},
		{
			CashflowSubGroup: constants.OTHER_CURRENT_ASSETS,
			CashflowGroup:    constants.CASHFLOW_GROUP_OPERATING,
			Category:         constants.CATEGORY_EQUITY,

			Name:      "Jasa Usaha",
			Type:      constants.TYPE_EQUITY,
			UserID:    userID,
			CompanyID: companyID,
			Code:      "33004",
		},
		{
			CashflowSubGroup: constants.EQUITY_CAPITAL,
			CashflowGroup:    constants.CASHFLOW_GROUP_FINANCING,
			Category:         constants.CATEGORY_EQUITY,

			Name:      "Jasa Modal",
			Type:      constants.TYPE_EQUITY,
			UserID:    userID,
			CompanyID: companyID,
			Code:      "33005",
		},
		{
			CashflowSubGroup: constants.OPERATIONAL_EXPENSES,
			CashflowGroup:    constants.CASHFLOW_GROUP_OPERATING,
			Category:         constants.CATEGORY_EQUITY,

			Name:      "Dana Pengurus",
			Type:      constants.TYPE_EQUITY,
			UserID:    userID,
			CompanyID: companyID,
			Code:      "33006",
		},
		{
			CashflowSubGroup: constants.OPERATIONAL_EXPENSES,
			CashflowGroup:    constants.CASHFLOW_GROUP_OPERATING,
			Category:         constants.CATEGORY_EQUITY,

			Name:      "Dana Lainnya",
			Type:      constants.TYPE_EQUITY,
			UserID:    userID,
			CompanyID: companyID,
			Code:      "33007",
		},

		{
			CashflowSubGroup: constants.OTHER_CURRENT_ASSETS,
			CashflowGroup:    constants.CASHFLOW_GROUP_OPERATING,
			Category:         constants.CATEGORY_DEBT,

			Name:      "Hutang SHU Ke Anggota",
			Type:      constants.TYPE_LIABILITY,
			UserID:    userID,
			CompanyID: companyID,
			Code:      "23004",
		},

		{
			CashflowSubGroup: constants.OPERATIONAL_EXPENSES,
			CashflowGroup:    constants.CASHFLOW_GROUP_OPERATING,
			Category:         constants.CATEGORY_EXPENSE,

			Name:      "Beban Sosial",
			Type:      constants.TYPE_EXPENSE,
			UserID:    userID,
			CompanyID: companyID,
			Code:      "53001",
		},
		{
			CashflowSubGroup: constants.OPERATIONAL_EXPENSES,
			CashflowGroup:    constants.CASHFLOW_GROUP_OPERATING,
			Category:         constants.CATEGORY_EXPENSE,

			Name:      "Beban Pendidikan",
			Type:      constants.TYPE_EXPENSE,
			UserID:    userID,
			CompanyID: companyID,
			Code:      "53001",
		},
		{
			CashflowSubGroup: constants.CASH_BANK,
			CashflowGroup:    constants.CASHFLOW_GROUP_CURRENT_ASSET,
			Category:         constants.CATEGORY_CURRENT_ASSET,

			Name:      "Kas Khusus Alokasi SHU",
			Type:      constants.TYPE_ASSET,
			UserID:    userID,
			CompanyID: companyID,
			Code:      "13001",
		},
		{
			CashflowSubGroup: constants.CASH_BANK,
			CashflowGroup:    constants.CASHFLOW_GROUP_CURRENT_ASSET,
			Category:         constants.CATEGORY_CURRENT_ASSET,

			Name:      "Kas Alokasi SHU Jasa Usaha",
			Type:      constants.TYPE_ASSET,
			UserID:    userID,
			CompanyID: companyID,
			Code:      "13002",
		},
		{
			CashflowSubGroup: constants.CASH_BANK,
			CashflowGroup:    constants.CASHFLOW_GROUP_CURRENT_ASSET,
			Category:         constants.CATEGORY_CURRENT_ASSET,

			Name:      "Kas Alokasi SHU Jasa Modal",
			Type:      constants.TYPE_ASSET,
			UserID:    userID,
			CompanyID: companyID,
			Code:      "13003",
		},
		{
			CashflowSubGroup: constants.CASH_BANK,
			CashflowGroup:    constants.CASHFLOW_GROUP_CURRENT_ASSET,
			Category:         constants.CATEGORY_CURRENT_ASSET,

			Name:      "Kas Alokasi SHU Dana Sosial",
			Type:      constants.TYPE_ASSET,
			UserID:    userID,
			CompanyID: companyID,
			Code:      "13004",
		},
		{
			CashflowSubGroup: constants.CASH_BANK,
			CashflowGroup:    constants.CASHFLOW_GROUP_CURRENT_ASSET,
			Category:         constants.CATEGORY_CURRENT_ASSET,

			Name:      "Kas Alokasi SHU Dana Pendidikan",
			Type:      constants.TYPE_ASSET,
			UserID:    userID,
			CompanyID: companyID,
			Code:      "13005",
		},
		{
			CashflowSubGroup: constants.CASH_BANK,
			CashflowGroup:    constants.CASHFLOW_GROUP_CURRENT_ASSET,
			Category:         constants.CATEGORY_CURRENT_ASSET,

			Name:      "Kas Alokasi SHU Dana Cadangan",
			Type:      constants.TYPE_ASSET,
			UserID:    userID,
			CompanyID: companyID,
			Code:      "13006",
		},
		{
			CashflowSubGroup: constants.CASH_BANK,
			CashflowGroup:    constants.CASHFLOW_GROUP_CURRENT_ASSET,
			Category:         constants.CATEGORY_CURRENT_ASSET,

			Name:      "Kas Alokasi SHU Dana Pengurus",
			Type:      constants.TYPE_ASSET,
			UserID:    userID,
			CompanyID: companyID,
			Code:      "13007",
		},
	}

}
