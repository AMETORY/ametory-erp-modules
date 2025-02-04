package transaction

import (
	"fmt"
	"net/http"
	"time"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/finance/account"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
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

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&models.TransactionModel{})
}

func (s *TransactionService) CreateTransaction(transaction *models.TransactionModel, amount float64) error {
	code := utils.RandString(10, false)
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
		var transSourceID, transDestID string = uuid.New().String(), uuid.New().String()
		if transaction.SourceID != nil {
			transaction.ID = transSourceID
			transaction.Code = code
			transaction.AccountID = transaction.SourceID
			transaction.Amount = amount
			transaction.TransactionRefID = &transDestID
			transaction.TransactionRefType = "transaction"
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
			transaction.ID = transDestID
			transaction.Code = code
			transaction.AccountID = transaction.DestinationID
			transaction.Amount = amount
			transaction.TransactionRefID = &transSourceID
			transaction.TransactionRefType = "transaction"
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

func (s *TransactionService) UpdateTransaction(id string, transaction *models.TransactionModel) error {
	return s.db.Where("id = ?", id).Updates(transaction).Error
}

func (s *TransactionService) DeleteTransaction(id string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var data models.TransactionModel
		err := tx.Where("id = ?", id).First(&data).Error
		if err != nil {
			return err
		}
		err = tx.Where("id = ?", id).Delete(&models.TransactionModel{}).Error
		if err != nil {
			return err
		}
		err = tx.Where("code = ?", data.Code).Delete(&models.TransactionModel{}).Error
		if err != nil {
			return err
		}
		return nil
	})
}

func (s *TransactionService) GetTransactionById(id string) (*models.TransactionModel, error) {
	var transaction models.TransactionModel
	err := s.db.Preload("Account").Select("transactions.*, accounts.name as account_name").Joins("LEFT JOIN accounts ON accounts.id = transactions.account_id").
		First(&transaction, "transactions.id = ?", id).Error
	return &transaction, err
}

func (s *TransactionService) GetTransactionByCode(code string) ([]models.TransactionModel, error) {
	var transaction []models.TransactionModel
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
	page := pg.With(stmt).Request(request).Response(new([]models.TransactionModel))
	page.Page = page.Page + 1
	return page, nil
}

func (s *TransactionService) GetByDateAndCompanyId(from, to time.Time, companyId string, page, limit int) ([]models.TransactionModel, error) {
	var transactions []models.TransactionModel
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
	stmt = stmt.Model(&models.TransactionModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.TransactionModel{})
	page.Page = page.Page + 1
	return page, nil
}

func (s *TransactionService) UpdateCreditDebit(transaction *models.TransactionModel, accountType models.AccountType) (*models.TransactionModel, error) {
	transaction.IsExpense = false
	transaction.IsIncome = false

	if accountType == models.EXPENSE {
		transaction.IsExpense = true
	}
	if accountType == models.REVENUE {
		transaction.IsIncome = true
	}
	switch accountType {
	case models.ASSET, models.EXPENSE, models.COST, models.CONTRA_LIABILITY, models.CONTRA_EQUITY, models.CONTRA_REVENUE:
		transaction.Debit = transaction.Amount
		transaction.Credit = 0
	case models.LIABILITY, models.EQUITY, models.REVENUE, models.CONTRA_ASSET, models.CONTRA_EXPENSE:
		transaction.Credit = transaction.Amount
		transaction.Debit = 0
	default:
		return transaction, fmt.Errorf("unhandled account type: %s", accountType)
	}

	return transaction, nil
}
