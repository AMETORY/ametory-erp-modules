package finance

import (
	"fmt"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/finance/account"
	"github.com/AMETORY/ametory-erp-modules/finance/transaction"
	"gorm.io/gorm"
)

type FinanceService struct {
	ctx                *context.ERPContext
	AccountService     *account.AccountService
	TransactionService *transaction.TransactionService
}

func NewFinanceService(ctx *context.ERPContext) *FinanceService {
	fmt.Println("INIT FINANCE SERVICE")
	var service = FinanceService{
		ctx: ctx,
	}
	service.AccountService = account.NewAccountService(ctx.DB, ctx)
	err := service.Migrate()
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return &service
}

func (s *FinanceService) Migrate() error {
	if s.ctx.SkipMigration {
		return nil
	}
	if err := account.Migrate(s.ctx.DB); err != nil {
		fmt.Println("ERROR ACCOUNT", err)
		return err
	}
	if err := transaction.Migrate(s.ctx.DB); err != nil {
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
