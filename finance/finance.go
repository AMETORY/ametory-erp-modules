package finance

import (
	"fmt"
	"log"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/finance/account"
	"github.com/AMETORY/ametory-erp-modules/finance/bank"
	"github.com/AMETORY/ametory-erp-modules/finance/report"
	"github.com/AMETORY/ametory-erp-modules/finance/transaction"
	"gorm.io/gorm"
)

type FinanceService struct {
	ctx                *context.ERPContext
	AccountService     *account.AccountService
	TransactionService *transaction.TransactionService
	BankService        *bank.BankService
	ReportService      *report.FinanceReportService
}

func NewFinanceService(ctx *context.ERPContext) *FinanceService {
	fmt.Println("INIT FINANCE SERVICE")
	var service = FinanceService{
		ctx: ctx,
	}
	service.AccountService = account.NewAccountService(ctx.DB, ctx)
	service.TransactionService = transaction.NewTransactionService(ctx.DB, ctx, service.AccountService)
	service.BankService = bank.NewBankService(ctx.DB, ctx)
	service.ReportService = report.NewFinanceReportService(ctx.DB, ctx, service.AccountService, service.TransactionService)
	err := service.Migrate()
	if err != nil {
		panic(err)
	}
	return &service
}

func (s *FinanceService) Migrate() error {
	if s.ctx.SkipMigration {
		return nil
	}
	if err := account.Migrate(s.ctx.DB); err != nil {
		log.Println("ERROR ACCOUNT MIGRATE", err)
		return err
	}
	if err := transaction.Migrate(s.ctx.DB); err != nil {
		log.Println("ERROR TRANSACTION MIGRATE", err)
		return err
	}
	// if err := transaction.Migrate(s.TransactionService.DB()); err != nil {
	// 	return err
	// }
	// if err := invoice.Migrate(s.InvoiceService.DB()); err != nil {
	// 	return err
	// }
	return nil
}
func (s *FinanceService) DB() *gorm.DB {
	return s.ctx.DB
}
