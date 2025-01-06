package transaction

import (
	"fmt"
	"net/http"
	"time"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/finance/account"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/google/uuid"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type TransactionService struct {
	db             *gorm.DB
	ctx            *context.ERPContext
	accountService *account.AccountService
}

func NewTransactionService(db *gorm.DB, ctx *context.ERPContext, accountService *account.AccountService) *TransactionService {
	return &TransactionService{db: db, ctx: ctx, accountService: accountService}
}

func (s *TransactionService) CreateTransaction(transaction *TransactionModel, amount float64) error {
	code := utils.RandString(10)
	if transaction.AccountID != nil {
		transaction.ID = uuid.New().String()
		transaction.Code = code
		transaction.Amount = amount
		account, err := s.accountService.GetAccountByID(*transaction.AccountID)
		if err != nil {
			return err
		}
		s.UpdateCreditDebit(transaction, account.Type)

		if err := s.db.Create(transaction).Error; err != nil {
			return err
		}
	} else {
		if transaction.SourceID != nil {
			transaction.ID = uuid.New().String()
			transaction.Code = code
			transaction.AccountID = transaction.SourceID
			transaction.Amount = amount
			account, err := s.accountService.GetAccountByID(*transaction.AccountID)
			if err != nil {
				return err
			}
			s.UpdateCreditDebit(transaction, account.Type)

			if err := s.db.Create(transaction).Error; err != nil {
				return err
			}
		}
		if transaction.DestinationID != nil {
			transaction.ID = uuid.New().String()
			transaction.Code = code
			transaction.AccountID = transaction.DestinationID
			transaction.Amount = amount
			account, err := s.accountService.GetAccountByID(*transaction.AccountID)
			if err != nil {
				return err
			}
			s.UpdateCreditDebit(transaction, account.Type)
			if err := s.db.Create(transaction).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *TransactionService) UpdateTransaction(id string, transaction *TransactionModel) error {
	return s.db.Where("id = ?", id).Updates(transaction).Error
}

func (s *TransactionService) DeleteTransaction(id string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var data TransactionModel
		err := tx.Where("id = ?", id).First(&data).Error
		if err != nil {
			return err
		}
		err = tx.Where("id = ?", id).Delete(&TransactionModel{}).Error
		if err != nil {
			return err
		}
		err = tx.Where("code = ?", data.Code).Delete(&TransactionModel{}).Error
		if err != nil {
			return err
		}
		return nil
	})
}

func (s *TransactionService) GetTransactionById(id string) (*TransactionModel, error) {
	var transaction TransactionModel
	err := s.db.Preload("Account").Select("transactions.*, accounts.name as account_name").Joins("LEFT JOIN accounts ON accounts.id = transactions.account_id").
		First(&transaction, "transactions.id = ?", id).Error
	return &transaction, err
}

func (s *TransactionService) GetTransactionByCode(code string) ([]TransactionModel, error) {
	var transaction []TransactionModel
	err := s.db.Preload("Account").Select("transactions.*, accounts.name as account_name").Joins("LEFT JOIN accounts ON accounts.id = transactions.account_id").
		First(&transaction, "transactions.code = ?", code).Error
	return transaction, err
}

func (s *TransactionService) GetTransactionByDate(from, to time.Time, request http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.Where("date BETWEEN ? AND ?", from, to)
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(new([]TransactionModel))
	page.Page = page.Page + 1
	return page, nil
}

func (s *TransactionService) GetByDateAndCompanyId(from, to time.Time, companyId string, page, limit int) ([]TransactionModel, error) {
	var transactions []TransactionModel
	err := s.db.Where("date BETWEEN ? AND ? AND company_id = ?", from, to, companyId).
		Offset((page - 1) * limit).Limit(limit).Find(&transactions).Error
	return transactions, err
}

func (s *TransactionService) GetTransactions(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.Preload("Account").Select("transactions.*, accounts.name as account_name").Joins("LEFT JOIN accounts ON accounts.id = transactions.account_id")
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	if search != "" {
		stmt = stmt.Where("accounts.name ILIKE ? OR accounts.code ILIKE ? OR transactions.code ILIKE ? OR transactions.description ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
			"%"+search+"%",
			"%"+search+"%",
		)
	}
	stmt = stmt.Model(&TransactionModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]TransactionModel{})
	page.Page = page.Page + 1
	return page, nil
}

func (s *TransactionService) UpdateCreditDebit(transaction *TransactionModel, accountType account.AccountType) (*TransactionModel, error) {
	transaction.IsExpense = false
	transaction.IsIncome = false

	if accountType == account.EXPENSE {
		transaction.IsExpense = true
	}
	if accountType == account.REVENUE {
		transaction.IsIncome = true
	}
	switch accountType {
	case account.ASSET, account.EXPENSE, account.COST, account.CONTRA_LIABILITY, account.CONTRA_EQUITY, account.CONTRA_REVENUE:
		transaction.Debit = transaction.Amount
		transaction.Credit = 0
	case account.LIABILITY, account.EQUITY, account.REVENUE, account.CONTRA_ASSET, account.CONTRA_EXPENSE:
		transaction.Credit = transaction.Amount
		transaction.Debit = 0
	default:
		return transaction, fmt.Errorf("unhandled account type: %s", accountType)
	}

	return transaction, nil
}
