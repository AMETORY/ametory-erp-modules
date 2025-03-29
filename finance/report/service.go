package report

import (
	"fmt"
	"net/http"
	"time"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/finance/account"
	"github.com/AMETORY/ametory-erp-modules/finance/transaction"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/morkid/paginate"
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

	var companyID *string
	if request.Header.Get("ID-Company") != "" {
		compID := request.Header.Get("ID-Company")
		companyID = &compID
	}
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

	// BEFORE
	pageBefore, err := s.transactionService.GetTransactionsByAccountID(accountID, nil, startDate, companyID, request)
	if err != nil {
		return nil, err
	}
	var balanceCurrent float64
	balanceBefore, balanceCurrent, _ := s.getBalance(pageBefore, &balanceCurrent)

	// CURRENT
	pageCurrent, err := s.transactionService.GetTransactionsByAccountID(accountID, startDate, endDate, companyID, request)
	if err != nil {
		return nil, err
	}
	balance, _, _ := s.getBalance(pageCurrent, &balanceCurrent)

	// AFTER
	pageAfter, err := s.transactionService.GetTransactionsByAccountID(accountID, endDate, nil, companyID, request)
	if err != nil {
		return nil, err
	}
	balanceAfter, _, _ := s.getBalance(pageAfter, &balanceCurrent)

	fmt.Println(balanceBefore, balance, balanceAfter)
	return &models.AccountReport{
		StartDate:      startDate,
		EndDate:        endDate,
		Account:        *account,
		BalanceBefore:  balanceBefore,
		TotalBalance:   balanceAfter,
		CurrentBalance: balance,
	}, nil
}
func (s *FinanceReportService) getBalance(page paginate.Page, currentBalance *float64) (float64, float64, float64) {

	items := page.Items.(*[]models.TransactionModel)
	newItems := make([]models.TransactionModel, 0)
	var balance, credit, debit float64
	for _, item := range *items {
		balance += item.Balance
		*currentBalance += item.Balance
		credit += item.Credit
		debit += item.Debit
		newItems = append(newItems, item)
	}

	page.Items = &newItems

	return balance, credit, debit
}
