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
		NetSurplusService:         net_surplus.NewNetSurplusService(ctx.DB, ctx, cooperativeSettingService, financeService),
		FinanceService:            financeService,
	}
	if err := service.Migrate(); err != nil {
		panic(err)
	}
	return &service
}

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
