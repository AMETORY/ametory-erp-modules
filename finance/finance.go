package finance

import (
	"fmt"

	"github.com/AMETORY/ametory-erp-modules/finance/account"
	"github.com/AMETORY/ametory-erp-modules/finance/transaction"
	"gorm.io/gorm"
)

type FinanceService struct {
	db                 *gorm.DB
	AccountService     *account.AccountService
	TransactionService *transaction.TransactionService
	SkipMigration      bool
}

func NewFinanceService(db *gorm.DB, skipMigrate bool) *FinanceService {
	fmt.Println("INIT FINANCE SERVICE")
	var service = FinanceService{
		db:                 db,
		SkipMigration:      skipMigrate,
		AccountService:     account.NewAccountService(db),
		TransactionService: transaction.NewTransactionService(db),
	}
	err := service.Migrate()
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return &service
}

func (s *FinanceService) Migrate() error {
	if s.SkipMigration {
		return nil
	}
	if err := account.Migrate(s.db); err != nil {
		fmt.Println("ERROR ACCOUNT", err)
		return err
	}
	if err := transaction.Migrate(s.db); err != nil {
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
	return s.db
}
