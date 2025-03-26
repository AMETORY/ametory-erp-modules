package cooperative

import (
	"github.com/AMETORY/ametory-erp-modules/company"
	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/cooperative/cooperative_member"
	"github.com/AMETORY/ametory-erp-modules/cooperative/cooperative_setting"
	"github.com/AMETORY/ametory-erp-modules/cooperative/loan_application"
	"github.com/AMETORY/ametory-erp-modules/cooperative/saving"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
)

type CooperativeService struct {
	ctx                       *context.ERPContext
	companyService            *company.CompanyService
	cooperativeSettingService *cooperative_setting.CooperativeSettingService
	cooperativeMemberService  *cooperative_member.CooperativeMemberService
	loanApplicationService    *loan_application.LoanApplicationService
	savingService             *saving.SavingService
}

func NewCooperativeService(ctx *context.ERPContext, companySrv *company.CompanyService) *CooperativeService {
	var service = CooperativeService{
		ctx:                       ctx,
		companyService:            companySrv,
		cooperativeSettingService: cooperative_setting.NewCooperativeSettingService(ctx.DB, ctx),
		cooperativeMemberService:  cooperative_member.NewCooperativeMemberService(ctx.DB, ctx),
		loanApplicationService:    loan_application.NewLoanApplicationService(ctx.DB, ctx),
		savingService:             saving.NewSavingService(ctx.DB, ctx),
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
	)
}
