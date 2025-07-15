package finance

import (
	"fmt"
	"log"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/finance/account"
	"github.com/AMETORY/ametory-erp-modules/finance/asset"
	"github.com/AMETORY/ametory-erp-modules/finance/bank"
	"github.com/AMETORY/ametory-erp-modules/finance/journal"
	"github.com/AMETORY/ametory-erp-modules/finance/report"
	"github.com/AMETORY/ametory-erp-modules/finance/tax"
	"github.com/AMETORY/ametory-erp-modules/finance/transaction"
	"gorm.io/gorm"
)

type FinanceService struct {
	ctx                *context.ERPContext
	AccountService     *account.AccountService
	TransactionService *transaction.TransactionService
	BankService        *bank.BankService
	JournalService     *journal.JournalService
	ReportService      *report.FinanceReportService
	TaxService         *tax.TaxService
	AssetService       *asset.AssetService
}

// NewFinanceService creates a new instance of FinanceService.
//
// The service is created by providing a pointer to a gorm.DB instance and a pointer to an ERPContext instance.
// The ERP context is used for authentication and authorization purposes, while the
// database instance is used for CRUD (Create, Read, Update, Delete) operations.
//
// The service will call the Migrate() method after creation to migrate the database.
// If the migration fails, the service will panic.
func NewFinanceService(ctx *context.ERPContext) *FinanceService {
	fmt.Println("INIT FINANCE SERVICE")
	var service = FinanceService{
		ctx: ctx,
	}
	service.AccountService = account.NewAccountService(ctx.DB, ctx)
	service.TransactionService = transaction.NewTransactionService(ctx.DB, ctx, service.AccountService)
	service.BankService = bank.NewBankService(ctx.DB, ctx)
	service.JournalService = journal.NewJournalService(ctx.DB, ctx, service.AccountService, service.TransactionService)
	service.ReportService = report.NewFinanceReportService(ctx.DB, ctx, service.AccountService, service.TransactionService)
	service.TaxService = tax.NewTaxService(ctx.DB, ctx, service.AccountService)
	service.AssetService = asset.NewAssetService(ctx.DB, ctx)
	err := service.Migrate()
	if err != nil {
		panic(err)
	}
	return &service
}

// Migrate migrates the database schema for the FinanceService.
//
// If the SkipMigration flag is true in the context, this method
// will not perform any migration and will return nil. Otherwise, it will
// attempt to auto-migrate the database to include the
// AccountModel, TransactionModel, JournalModel, TaxModel, and AssetModel schemas.
// If the migration process encounters an error, it will return that error.
// Otherwise, it will return nil upon successful migration.
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
	if err := journal.Migrate(s.ctx.DB); err != nil {
		log.Println("ERROR JOURNAL MIGRATE", err)
		return err
	}
	if err := tax.Migrate(s.ctx.DB); err != nil {
		log.Println("ERROR TAX MIGRATE", err)
		return err
	}
	if err := report.Migrate(s.ctx.DB); err != nil {
		log.Println("ERROR REPORT MIGRATE", err)
		return err
	}
	if err := asset.Migrate(s.ctx.DB); err != nil {
		log.Println("ERROR ASSET MIGRATE", err)
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

// DB returns the underlying GORM database connection used by the FinanceService.
//
// This method provides access to the database instance associated with the
// current ERP context, enabling CRUD operations within the service.

func (s *FinanceService) DB() *gorm.DB {
	return s.ctx.DB
}
