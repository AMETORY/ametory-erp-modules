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
			account, err := s.accountService.GetAccountByID(*transaction.SourceID)
			if err != nil {
				return err
			}
			s.UpdateCreditDebit(transaction, account.Type)
			if transaction.IsTransfer {
				transaction.Credit = amount
				transaction.Debit = 0
			}

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
			account, err := s.accountService.GetAccountByID(*transaction.DestinationID)
			if err != nil {
				return err
			}
			s.UpdateCreditDebit(transaction, account.Type)
			if transaction.IsTransfer {
				transaction.Debit = amount
				transaction.Credit = 0
				transaction.IsTransfer = false
			}
			if account.Type == models.ASSET {
				transaction.IsIncome = false
				transaction.IsExpense = false
			}

			if err := s.db.Create(transaction).Error; err != nil {
				return err
			}

		}
	}

	return nil
}

func (s *TransactionService) UpdateTransaction(id string, transaction *models.TransactionModel) error {
	// return s.db.Where("id = ?", id).Updates(transaction).Error
	return s.db.Transaction(func(tx *gorm.DB) error {
		if transaction.Debit > 0 {
			transaction.Debit = transaction.Amount
		}
		if transaction.Credit > 0 {
			transaction.Credit = transaction.Amount
		}
		err := tx.Model(&models.TransactionModel{}).Where("id = ?", id).Updates(transaction).Error
		if err != nil {
			return err
		}
		var trans2 models.TransactionModel
		err = tx.Where("code = ? and id != ?", transaction.Code, transaction.ID).First(&trans2).Error
		if err == nil {
			var credit, debit float64 = transaction.Debit, transaction.Credit
			if credit > 0 {
				credit = transaction.Amount
			}
			if debit > 0 {
				debit = transaction.Amount
			}
			err = tx.Model(&models.TransactionModel{}).Where("id = ?", trans2.ID).Updates(map[string]any{
				"credit":      credit,
				"debit":       debit,
				"description": transaction.Description,
				"date":        transaction.Date,
				"amount":      transaction.Amount,
			}).Error
			if err != nil {
				return err
			}
		}

		return nil
	})
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

func (s *TransactionService) GetTransactionsByAccountID(accountID string, startDate *time.Time, endDate *time.Time, companyID *string, request http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.Preload("Account").Select("transactions.*, accounts.name as account_name").Joins("LEFT JOIN accounts ON accounts.id = transactions.account_id")
	if companyID != nil {
		stmt = stmt.Where("transactions.company_id = ?", *companyID)
	}
	if startDate != nil {
		stmt = stmt.Where("transactions.date >= ?", *startDate)
	}
	if endDate != nil {
		stmt = stmt.Where("transactions.date < ?", *endDate)
	}
	stmt = stmt.Where("transactions.account_id = ?", accountID)
	stmt = stmt.Model(&models.TransactionModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.TransactionModel{})
	page.Page = page.Page + 1
	items := page.Items.(*[]models.TransactionModel)
	newItems := make([]models.TransactionModel, 0)
	for _, item := range *items {
		if item.TransactionRefID != nil {
			var transRef models.TransactionModel
			err := s.db.Preload("Account").Where("id = ?", item.TransactionRefID).First(&transRef).Error
			if err == nil {
				item.TransactionRef = &transRef
			}
		}
		item.Balance = s.getBalance(item)
		newItems = append(newItems, item)
	}
	page.Items = &newItems
	return page, nil
}
func (s *TransactionService) GetTransactions(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.Preload("Account").Select("transactions.*, accounts.name as account_name").Joins("LEFT JOIN accounts ON accounts.id = transactions.account_id")
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("transactions.company_id = ?", request.Header.Get("ID-Company"))
	}
	if search != "" {
		stmt = stmt.Where("accounts.name ILIKE ? OR accounts.code ILIKE ? OR transactions.code ILIKE ? OR transactions.description ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
			"%"+search+"%",
			"%"+search+"%",
		)
	}

	switch request.URL.Query().Get("type") {
	case "INCOME":
		stmt = stmt.Where("transactions.is_income = ?", true)
	case "EXPENSE":
		stmt = stmt.Where("transactions.is_expense = ?", true)
	case "EQUITY":
		stmt = stmt.Where("transactions.is_equity = ?", true)
	case "TRANSFER":
		stmt = stmt.Where("transactions.is_transfer = ?", true)
	}

	if request.URL.Query().Get("account_id") != "" {
		stmt = stmt.Where("transactions.account_id = ?", request.URL.Query().Get("account_id"))
	}

	stmt = stmt.Model(&models.TransactionModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.TransactionModel{})
	page.Page = page.Page + 1
	items := page.Items.(*[]models.TransactionModel)
	newItems := make([]models.TransactionModel, 0)
	for _, item := range *items {
		if item.TransactionRefID != nil {
			var transRef models.TransactionModel
			err := s.db.Preload("Account").Where("id = ?", item.TransactionRefID).First(&transRef).Error
			if err == nil {
				item.TransactionRef = &transRef
			}
		}
		item.Balance = s.getBalance(item)
		newItems = append(newItems, item)
	}
	page.Items = &newItems
	return page, nil
}

func (s *TransactionService) UpdateCreditDebit(transaction *models.TransactionModel, accountType models.AccountType) (*models.TransactionModel, error) {
	// transaction.IsExpense = false
	// transaction.IsIncome = false

	if accountType == models.EXPENSE || accountType == models.COST {
		transaction.IsExpense = true
	}
	if accountType == models.REVENUE || accountType == models.INCOME {
		transaction.IsIncome = true
	}
	if accountType == models.EQUITY {
		transaction.IsEquity = true
	}
	switch accountType {
	case models.EXPENSE, models.COST, models.CONTRA_LIABILITY, models.CONTRA_EQUITY, models.CONTRA_REVENUE:
		transaction.Debit = transaction.Amount
		transaction.Credit = 0
	case models.LIABILITY, models.EQUITY, models.REVENUE, models.INCOME, models.CONTRA_ASSET, models.CONTRA_EXPENSE:
		transaction.Credit = transaction.Amount
		transaction.Debit = 0
	case models.ASSET:
		if transaction.IsIncome {
			transaction.Debit = transaction.Amount
			transaction.Credit = 0
		}
		if transaction.IsEquity {
			transaction.Debit = transaction.Amount
			transaction.Credit = 0
			transaction.IsEquity = false
		}
		if transaction.IsExpense {
			transaction.Credit = transaction.Amount
			transaction.Debit = 0
		}
	default:
		return transaction, fmt.Errorf("unhandled account type: %s", accountType)
	}

	fmt.Printf("account type: %v, is income: %v, is expense: %v, is equity: %v\n", accountType, transaction.IsIncome, transaction.IsExpense, transaction.IsEquity)

	return transaction, nil
}

func (s *TransactionService) getBalance(transaction models.TransactionModel) float64 {
	switch transaction.Account.Type {
	case models.EXPENSE, models.COST, models.CONTRA_LIABILITY, models.CONTRA_EQUITY, models.CONTRA_REVENUE:
		return transaction.Debit - transaction.Credit
	case models.LIABILITY, models.EQUITY, models.REVENUE, models.INCOME, models.CONTRA_ASSET, models.CONTRA_EXPENSE:
		return transaction.Credit - transaction.Debit
	case models.ASSET:
		return transaction.Debit - transaction.Credit
	}
	return 0
}
