package report

import (
	"fmt"
	"net/http"
	"time"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/finance/account"
	"github.com/AMETORY/ametory-erp-modules/finance/transaction"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"gorm.io/gorm"
)

type FinanceReportService struct {
	db                 *gorm.DB
	ctx                *context.ERPContext
	accountService     *account.AccountService
	transactionService *transaction.TransactionService
}

func NewFinanceReportService(db *gorm.DB, ctx *context.ERPContext, accountService *account.AccountService, transactionService *transaction.TransactionService) *FinanceReportService {
	return &FinanceReportService{
		db:                 db,
		ctx:                ctx,
		accountService:     accountService,
		transactionService: transactionService,
	}
}

func (s *FinanceReportService) GenerateProfitLoss(report *models.ProfitLoss) error {

	return nil
}

func (s *FinanceReportService) GenerateAccountReport(accountID string, request http.Request) (*models.AccountReport, error) {
	account, err := s.accountService.GetAccountByID(accountID)
	if err != nil {
		return nil, err
	}
	if request.URL.Query().Get("start_date") == "" {
		return nil, fmt.Errorf("start date is required")
	}
	if request.URL.Query().Get("end_date") == "" {
		return nil, fmt.Errorf("end date is required")
	}

	fmt.Println(account)

	// var companyID *string
	// if request.Header.Get("ID-Company") != "" {
	// 	compID := request.Header.Get("ID-Company")
	// 	companyID = &compID
	// }
	var startDate, endDate *time.Time
	startDateParsed, err := time.Parse("2006-01-02", request.URL.Query().Get("start_date"))
	if err != nil {
		return nil, err
	}
	endDateParsed, err := time.Parse("2006-01-02", request.URL.Query().Get("end_date"))
	if err != nil {
		return nil, err
	}
	startDate = &startDateParsed
	endDate = &endDateParsed

	var balanceCurrent, balanceBefore float64
	// BEFORE

	debit, credit, _ := s.GetAccountBalance(accountID, nil, startDate)
	switch account.Type {
	case models.EXPENSE, models.COST, models.CONTRA_LIABILITY, models.CONTRA_EQUITY, models.CONTRA_REVENUE:
		balanceCurrent = debit - credit
	case models.LIABILITY, models.EQUITY, models.REVENUE, models.INCOME, models.CONTRA_ASSET, models.CONTRA_EXPENSE:
		balanceCurrent = credit - debit
	case models.ASSET:
		balanceCurrent = debit - credit
	}
	balanceBefore = balanceCurrent

	// CURRENT
	pageCurrent, err := s.GetAccountTransactions(accountID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	balance, _, _ := s.getBalance(&pageCurrent, &balanceCurrent)

	var balanceAfter float64
	// AFTER
	debit, credit, _ = s.GetAccountBalance(accountID, endDate, nil)
	switch account.Type {
	case models.EXPENSE, models.COST, models.CONTRA_LIABILITY, models.CONTRA_EQUITY, models.CONTRA_REVENUE:
		balanceAfter = debit - credit
	case models.LIABILITY, models.EQUITY, models.REVENUE, models.INCOME, models.CONTRA_ASSET, models.CONTRA_EXPENSE:
		balanceAfter = credit - debit
	case models.ASSET:
		balanceAfter = debit - credit
	}

	return &models.AccountReport{
		StartDate:      startDate,
		EndDate:        endDate,
		Account:        *account,
		BalanceBefore:  balanceBefore,
		TotalBalance:   balanceBefore + balance + balanceAfter,
		CurrentBalance: balance,
		Transactions:   pageCurrent,
	}, nil
}
func (s *FinanceReportService) getBalance(page *[]models.TransactionModel, currentBalance *float64) (float64, float64, float64) {

	newItems := make([]models.TransactionModel, 0)
	var balance, credit, debit float64
	for _, item := range *page {
		if item.TransactionRefID != nil {
			var transRef models.TransactionModel
			err := s.db.Preload("Account").Where("id = ?", item.TransactionRefID).First(&transRef).Error
			if err == nil {
				item.TransactionRef = &transRef
			}
		}
		curBalance := s.getBalanceAmount(item)
		balance += curBalance
		// fmt.Printf("balance %f, currentBalance %f\n", balance, *currentBalance)
		*currentBalance += curBalance
		item.Balance = *currentBalance
		credit += item.Credit
		debit += item.Debit

		newItems = append(newItems, item)
	}

	*page = newItems

	return balance, credit, debit
}

func (s *FinanceReportService) GetAccountBalance(accountID string, startDate *time.Time, endDate *time.Time) (float64, float64, error) {
	amount := struct {
		Credit float64 `sql:"credit"`
		Debit  float64 `sql:"debit"`
	}{}
	db := s.db.Model(&models.TransactionModel{}).Select("sum(credit) as credit, sum(debit) as debit").Where("account_id = ?", accountID)
	if startDate != nil {
		db = db.Where("date >= ?", startDate)
	}
	if endDate != nil {
		db = db.Where("date < ?", endDate)
	}
	err := db.Scan(&amount).Error
	if err != nil {
		return 0, 0, err
	}

	return amount.Debit, amount.Credit, nil
}

func (s *FinanceReportService) GetAccountTransactions(accountID string, startDate *time.Time, endDate *time.Time) ([]models.TransactionModel, error) {
	var transactions []models.TransactionModel
	db := s.db.Preload("Account").Select("transactions.*, accounts.name as account_name").Joins("LEFT JOIN accounts ON accounts.id = transactions.account_id")

	if startDate != nil {
		db = db.Where("transactions.date >= ?", *startDate)
	}
	if endDate != nil {
		db = db.Where("transactions.date < ?", *endDate)
	}
	db = db.Where("transactions.account_id = ?", accountID)

	db = db.Order("date asc")
	err := db.Find(&transactions).Error
	if err != nil {
		return nil, err
	}

	return transactions, nil

}

func (s *FinanceReportService) getBalanceAmount(transaction models.TransactionModel) float64 {
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
