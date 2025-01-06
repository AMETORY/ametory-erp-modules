package transaction

import (
	"fmt"
	"net/http"
	"time"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/finance/account"
	"github.com/AMETORY/ametory-erp-modules/utils"
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

	if transaction.SourceID != nil {
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
	return nil
}

func (s *TransactionService) UpdateTransaction(transaction *TransactionModel) error {
	return s.db.Save(transaction).Error
}

func (s *TransactionService) DeleteTransaction(transaction *TransactionModel) error {
	return s.db.Delete(transaction).Error
}

func (s *TransactionService) GetTransactionById(id string) (*TransactionModel, error) {
	var transaction TransactionModel
	err := s.db.First(&transaction, "id = ?", id).Error
	return &transaction, err
}

func (s *TransactionService) GetTransactionByCode(code string) (*TransactionModel, error) {
	var transaction TransactionModel
	err := s.db.First(&transaction, "code = ?", code).Error
	return &transaction, err
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
	stmt := s.db.Joins("LEFT JOIN accounts", "accounts.id = transactions.account_id")
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	if search != "" {
		stmt = stmt.Where("accounts.name LIKE ? OR accounts.code LIKE ? OR transactions.code LIKE ? OR transactions.description LIKE ?",
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
