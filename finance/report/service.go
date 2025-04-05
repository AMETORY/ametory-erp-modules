package report

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/finance/account"
	"github.com/AMETORY/ametory-erp-modules/finance/transaction"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
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

	// fmt.Println(account)

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
			if item.TransactionRefType == "journal" {
				var journalRef models.JournalModel
				err := s.db.Where("id = ?", item.TransactionRefID).First(&journalRef).Error
				if err == nil {
					item.JournalRef = &journalRef
				}
			}
			if item.TransactionRefType == "transaction" {
				var transRef models.TransactionModel
				err := s.db.Preload("Account").Where("id = ?", item.TransactionRefID).First(&transRef).Error
				if err == nil {
					item.TransactionRef = &transRef
				}
			}
			if item.TransactionRefType == "sales" {
				var salesRef models.SalesModel
				err := s.db.Where("id = ?", item.TransactionRefID).First(&salesRef).Error
				if err == nil {
					item.SalesRef = &salesRef
				}
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
	case models.EXPENSE, models.COST, models.CONTRA_LIABILITY, models.CONTRA_EQUITY, models.CONTRA_REVENUE, models.RECEIVABLE:
		return transaction.Debit - transaction.Credit
	case models.LIABILITY, models.EQUITY, models.REVENUE, models.INCOME, models.CONTRA_ASSET, models.CONTRA_EXPENSE:
		return transaction.Credit - transaction.Debit
	case models.ASSET:
		return transaction.Debit - transaction.Credit
	}
	return 0
}

func (s *FinanceReportService) GenerateCogsReport(report models.GeneralReport) (*models.COGSReport, error) {
	var inventoryAccount models.AccountModel
	err := s.db.Where("is_inventory_account = ? and company_id = ?", true, report.CompanyID).First(&inventoryAccount).Error
	if err != nil {
		return nil, errors.New("inventory account not found")
	}

	var beginningInventory, purchases, freightInAndOtherCost, totalPurchases, purchaseReturns, purchaseDiscounts, totalPurchaseDiscounts, netPurchases, goodsAvailable, endingInventory, cogs float64
	amount := struct {
		Sum float64 `sql:"sum"`
	}{}
	err = s.db.Model(&models.TransactionModel{}).
		Where("date < ?", report.StartDate).
		Select("sum(debit-credit) as sum").
		Where("account_id = ?", inventoryAccount.ID).
		Scan(&amount).Error
	if err != nil {
		return nil, err
	}
	beginningInventory = amount.Sum

	err = s.db.Model(&models.TransactionModel{}).
		Where("is_purchase_cost = ?", false).
		Where("debit > ?", 0).
		Where("date between ? and ?", report.StartDate, report.EndDate).
		Select("sum(debit-credit) as sum").
		Where("account_id = ?", inventoryAccount.ID).
		Scan(&amount).Error
	if err != nil {
		return nil, err
	}
	purchases = amount.Sum

	err = s.db.Model(&models.TransactionModel{}).
		Where("is_purchase_cost = ?", true).
		Where("debit > ?", 0).
		Where("date between ? and ?", report.StartDate, report.EndDate).
		Select("sum(debit-credit) as sum").
		Where("account_id = ?", inventoryAccount.ID).
		Scan(&amount).Error
	if err != nil {
		return nil, err
	}
	freightInAndOtherCost = amount.Sum
	totalPurchases = purchases + freightInAndOtherCost

	err = s.db.Model(&models.TransactionModel{}).
		Where("is_return = ?", true).
		Where("date between ? and ?", report.StartDate, report.EndDate).
		Select("sum(debit-credit) as sum").
		Where("account_id = ?", inventoryAccount.ID).
		Scan(&amount).Error
	if err != nil {
		return nil, err
	}
	purchaseReturns = amount.Sum
	err = s.db.Model(&models.TransactionModel{}).
		Where("is_discount = ?", true).
		Where("date between ? and ?", report.StartDate, report.EndDate).
		Select("sum(debit-credit) as sum").
		Where("account_id = ?", inventoryAccount.ID).
		Scan(&amount).Error
	if err != nil {
		return nil, err
	}
	purchaseDiscounts = amount.Sum

	totalPurchaseDiscounts = purchaseReturns + purchaseDiscounts

	err = s.db.Model(&models.TransactionModel{}).
		Where("date < ?", report.EndDate).
		Select("sum(debit-credit) as sum").
		Where("account_id = ?", inventoryAccount.ID).
		Scan(&amount).Error
	if err != nil {
		return nil, err
	}
	endingInventory = amount.Sum

	netPurchases = totalPurchases + totalPurchaseDiscounts
	goodsAvailable = beginningInventory + netPurchases
	cogs = goodsAvailable - endingInventory

	cogsData := models.COGSReport{
		BeginningInventory:     beginningInventory,
		Purchases:              purchases,
		FreightInAndOtherCost:  freightInAndOtherCost,
		TotalPurchases:         totalPurchases,
		PurchaseReturns:        purchaseReturns,
		PurchaseDiscounts:      purchaseDiscounts,
		TotalPurchaseDiscounts: totalPurchaseDiscounts,
		NetPurchases:           netPurchases,
		GoodsAvailable:         goodsAvailable,
		EndingInventory:        endingInventory,
		COGS:                   cogs,
	}
	cogsData.StartDate = report.StartDate
	cogsData.EndDate = report.EndDate
	utils.LogJson(cogsData)
	return &cogsData, nil
}

func (s *FinanceReportService) GenerateProfitLossReport(report models.GeneralReport) (*models.ProfitLossReport, error) {
	profitLoss := models.ProfitLossReport{}
	cogsReport, err := s.GenerateCogsReport(report)
	if err != nil {
		return nil, err
	}

	revenueAccounts := []models.AccountModel{}
	err = s.db.Where("type IN (?)", []models.AccountType{models.INCOME, models.REVENUE}).Find(&revenueAccounts).Error
	if err != nil {
		return nil, err
	}
	revenueSum := 0.0
	for _, revenue := range revenueAccounts {
		amount := struct {
			Sum float64 `sql:"sum"`
		}{}
		err = s.db.Model(&models.TransactionModel{}).
			Where("date between ? and ?", report.StartDate, report.EndDate).
			Select("sum(credit-debit) as sum").
			Joins("JOIN accounts ON accounts.id = transactions.account_id").
			Where("transactions.account_id = ?", revenue.ID).
			Scan(&amount).Error
		if err != nil {
			return nil, err
		}
		profitLoss.Profit = append(profitLoss.Profit, models.ProfilLossAccount{
			ID:   revenue.ID,
			Name: revenue.Name,
			Code: revenue.Code,
			Sum:  amount.Sum,
		})
		revenueSum += amount.Sum
	}

	profitLoss.Profit = append(profitLoss.Profit, models.ProfilLossAccount{
		Name: "Harga Pokok Penjualan",
		Sum:  cogsReport.COGS,
	})

	profitLoss.GrossProfit = revenueSum - cogsReport.COGS

	expenseAccounts := []models.AccountModel{}
	err = s.db.Where("type IN (?)", []models.AccountType{models.EXPENSE}).Find(&expenseAccounts).Error
	if err != nil {
		return nil, err
	}
	expenseSum := 0.0
	for _, expense := range expenseAccounts {
		amount := struct {
			Sum float64 `sql:"sum"`
		}{}
		err = s.db.Model(&models.TransactionModel{}).
			Where("date between ? and ?", report.StartDate, report.EndDate).
			Select("sum(debit-credit) as sum").
			Joins("JOIN accounts ON accounts.id = transactions.account_id").
			Where("transactions.account_id = ?", expense.ID).
			Scan(&amount).Error
		if err != nil {
			return nil, err
		}
		profitLoss.Loss = append(profitLoss.Loss, models.ProfilLossAccount{
			ID:   expense.ID,
			Name: expense.Name,
			Code: expense.Code,
			Sum:  amount.Sum,
		})
		expenseSum += amount.Sum
	}

	profitLoss.TotalExpense = expenseSum
	profitLoss.NetProfit = profitLoss.GrossProfit - profitLoss.TotalExpense
	return &profitLoss, nil
}
