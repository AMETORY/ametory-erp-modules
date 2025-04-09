package report

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/finance/account"
	"github.com/AMETORY/ametory-erp-modules/finance/transaction"
	"github.com/AMETORY/ametory-erp-modules/shared/constants"
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
			if item.TransactionSecondaryRefType == "sales" {
				var salesRef models.SalesModel
				err := s.db.Where("id = ?", item.TransactionSecondaryRefID).First(&salesRef).Error
				if err == nil {
					item.SalesRef = &salesRef
				}
			}
			if item.TransactionRefType == "purchase" {
				var purchaseRef models.PurchaseOrderModel
				err := s.db.Where("id = ?", item.TransactionRefID).First(&purchaseRef).Error
				if err == nil {
					item.PurchaseRef = &purchaseRef
				}
			}
			if item.TransactionSecondaryRefType == "purchase" {
				var purchaseRef models.PurchaseOrderModel
				err := s.db.Where("id = ?", item.TransactionSecondaryRefID).First(&purchaseRef).Error
				if err == nil {
					item.PurchaseRef = &purchaseRef
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
		Where("is_purchase = ?", true).
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
		Select("sum(credit-debit) as sum").
		Where("account_id = ?", inventoryAccount.ID).
		Scan(&amount).Error
	if err != nil {
		return nil, err
	}
	purchaseReturns = amount.Sum
	err = s.db.Model(&models.TransactionModel{}).
		Where("is_discount = ?", true).
		Where("date between ? and ?", report.StartDate, report.EndDate).
		Select("sum(credit-debit) as sum").
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

	netPurchases = totalPurchases - totalPurchaseDiscounts
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
		InventoryAccount:       inventoryAccount,
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
	err = s.db.Where("type IN (?)", []models.AccountType{models.INCOME, models.REVENUE, models.CONTRA_REVENUE}).Find(&revenueAccounts).Error
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
		profitLoss.Profit = append(profitLoss.Profit, models.ProfitLossAccount{
			ID:   revenue.ID,
			Name: revenue.Name,
			Code: revenue.Code,
			Sum:  amount.Sum,
		})
		revenueSum += amount.Sum
	}

	profitLoss.Profit = append(profitLoss.Profit, models.ProfitLossAccount{
		Name: "Harga Pokok Penjualan",
		Sum:  -cogsReport.COGS,
		Link: "/cogs",
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
		profitLoss.Loss = append(profitLoss.Loss, models.ProfitLossAccount{
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

func (s *FinanceReportService) GenerateBalanceSheet(report models.GeneralReport) (*models.BalanceSheet, error) {
	balanceSheet := models.BalanceSheet{}
	balanceSheet.StartDate = report.StartDate
	balanceSheet.EndDate = report.EndDate

	// ASSETS
	// FIXED ACCOUNT
	fixedAccounts := []models.AccountModel{}
	err := s.db.Where("type = ? AND cashflow_group = ? AND company_id = ?", "ASSET", "fixed_asset", report.CompanyID).Find(&fixedAccounts).Error
	if err != nil {
		return nil, errors.New("fixedAccounts account not found")
	}
	fixedAmount := 0.0
	for _, expense := range fixedAccounts {
		amount := struct {
			Sum float64 `sql:"sum"`
		}{}
		err = s.db.Model(&models.TransactionModel{}).
			Where("date <  ?", report.EndDate).
			Select("sum(debit-credit) as sum").
			Joins("JOIN accounts ON accounts.id = transactions.account_id").
			Where("transactions.account_id = ?", expense.ID).
			Scan(&amount).Error
		if err != nil {
			return nil, err
		}
		balanceSheet.FixedAssets = append(balanceSheet.FixedAssets, models.BalanceSheetAccount{
			ID:   expense.ID,
			Name: expense.Name,
			Code: expense.Code,
			Sum:  amount.Sum,
		})
		fixedAmount += amount.Sum
	}
	balanceSheet.TotalFixed = fixedAmount

	// CURRENT ACCOUNT
	currentAccounts := []models.AccountModel{}
	err = s.db.Where("type = ? AND cashflow_group = ? AND company_id = ?", "ASSET", "current_asset", report.CompanyID).Find(&currentAccounts).Error
	if err != nil {

	}
	currentAmount := 0.0
	for _, expense := range currentAccounts {
		amount := struct {
			Sum float64 `sql:"sum"`
		}{}
		err = s.db.Model(&models.TransactionModel{}).
			Where("date <  ?", report.EndDate).
			Select("sum(debit-credit) as sum").
			Joins("JOIN accounts ON accounts.id = transactions.account_id").
			Where("transactions.account_id = ?", expense.ID).
			Scan(&amount).Error
		if err != nil {
			return nil, err
		}
		balanceSheet.CurrentAssets = append(balanceSheet.CurrentAssets, models.BalanceSheetAccount{
			ID:   expense.ID,
			Name: expense.Name,
			Code: expense.Code,
			Sum:  amount.Sum,
		})
		currentAmount += amount.Sum
	}

	// RECEIVABLE ACCOUNT
	receivableAccounts := []models.AccountModel{}
	err = s.db.Where("type = ?  AND company_id = ?", "RECEIVABLE", report.CompanyID).Find(&receivableAccounts).Error
	if err != nil {
		return nil, errors.New("receivableAccounts account not found")
	}

	for _, expense := range receivableAccounts {
		amount := struct {
			Sum float64 `sql:"sum"`
		}{}
		err = s.db.Model(&models.TransactionModel{}).
			Where("date <  ?", report.EndDate).
			Select("sum(debit-credit) as sum").
			Joins("JOIN accounts ON accounts.id = transactions.account_id").
			Where("transactions.account_id = ?", expense.ID).
			Scan(&amount).Error
		if err != nil {
			return nil, err
		}
		balanceSheet.CurrentAssets = append(balanceSheet.CurrentAssets, models.BalanceSheetAccount{
			ID:   expense.ID,
			Name: expense.Name,
			Code: expense.Code,
			Sum:  amount.Sum,
		})
		currentAmount += amount.Sum
	}

	// INVENTORY
	report.StartDate = time.Time{}
	cogsReport, err := s.GenerateCogsReport(report)
	if err != nil {
		return nil, err
	}
	balanceSheet.CurrentAssets = append(balanceSheet.CurrentAssets, models.BalanceSheetAccount{
		ID:   cogsReport.InventoryAccount.ID,
		Code: cogsReport.InventoryAccount.Code,
		Name: cogsReport.InventoryAccount.Name,
		Sum:  cogsReport.EndingInventory,
	})

	currentAmount += cogsReport.EndingInventory
	balanceSheet.TotalCurrent = currentAmount

	balanceSheet.TotalAssets = balanceSheet.TotalFixed + balanceSheet.TotalCurrent
	// LIABILITY AND EQUITY

	// LIABILITY ACCOUNT
	liabilityAccounts := []models.AccountModel{}
	err = s.db.Where("type = ?  AND company_id = ?", "LIABILITY", report.CompanyID).Find(&liabilityAccounts).Error
	if err != nil {
		return nil, errors.New("liabilityAccounts account not found")
	}
	liabilityAmount := 0.0
	for _, expense := range liabilityAccounts {
		amount := struct {
			Sum float64 `sql:"sum"`
		}{}
		err = s.db.Model(&models.TransactionModel{}).
			Where("date <  ?", report.EndDate).
			Select("sum(credit-debit) as sum").
			Joins("JOIN accounts ON accounts.id = transactions.account_id").
			Where("transactions.account_id = ?", expense.ID).
			Scan(&amount).Error
		if err != nil {
			return nil, err
		}
		balanceSheet.LiableAssets = append(balanceSheet.LiableAssets, models.BalanceSheetAccount{
			ID:   expense.ID,
			Name: expense.Name,
			Code: expense.Code,
			Sum:  amount.Sum,
		})
		liabilityAmount += amount.Sum
	}

	balanceSheet.TotalLiability = liabilityAmount

	// EQUITY ACCOUNT
	equityAccounts := []models.AccountModel{}
	err = s.db.Where("type = ?  AND company_id = ?", "EQUITY", report.CompanyID).Find(&equityAccounts).Error
	if err != nil {
		return nil, errors.New("equityAccounts account not found")
	}
	equityAmount := 0.0
	for _, expense := range equityAccounts {
		amount := struct {
			Sum float64 `sql:"sum"`
		}{}
		err = s.db.Model(&models.TransactionModel{}).
			Where("date <  ?", report.EndDate).
			Select("sum(credit-debit) as sum").
			Joins("JOIN accounts ON accounts.id = transactions.account_id").
			Where("transactions.account_id = ?", expense.ID).
			Scan(&amount).Error
		if err != nil {
			return nil, err
		}
		balanceSheet.Equity = append(balanceSheet.Equity, models.BalanceSheetAccount{
			ID:   expense.ID,
			Name: expense.Name,
			Code: expense.Code,
			Sum:  amount.Sum,
		})
		equityAmount += amount.Sum
	}

	profitLoss, err := s.GenerateProfitLossReport(report)
	if err != nil {
		return nil, err
	}

	// PROFIT AND LOSS
	balanceSheet.Equity = append(balanceSheet.Equity, models.BalanceSheetAccount{
		Name: "Laba Ditahan",
		Sum:  profitLoss.NetProfit,
		Link: "/profit-loss-statement",
	})
	equityAmount += profitLoss.NetProfit
	balanceSheet.TotalEquity = equityAmount
	balanceSheet.TotalLiabilitiesAndEquity = balanceSheet.TotalLiability + balanceSheet.TotalEquity

	return &balanceSheet, nil
}

func (s *FinanceReportService) GenerateCapitalChangeReport(report models.GeneralReport) (*models.CapitalChangeReport, error) {
	capitalChange := models.CapitalChangeReport{}
	equityAccounts := []models.AccountModel{}
	err := s.db.Where("type = ?  AND company_id = ?", "EQUITY", report.CompanyID).Find(&equityAccounts).Error
	if err != nil {
		return nil, errors.New("equityAccounts account not found")
	}
	// Opening Balance
	openingBalance := 0.0
	for _, v := range equityAccounts {
		amount := struct {
			Sum float64 `sql:"sum"`
		}{}
		err := s.db.Model(&models.TransactionModel{}).
			Where("date <  ?", report.EndDate).
			Select("sum(credit) as sum").
			Joins("JOIN accounts ON accounts.id = transactions.account_id").
			Where("transactions.account_id = ?", v.ID).
			Where("is_opening_balance = ?", true).
			Scan(&amount).Error
		if err != nil {
			return nil, err
		}

		openingBalance += amount.Sum
	}
	profitLoss, err := s.GenerateProfitLossReport(report)
	if err != nil {
		return nil, err
	}

	profitLossBalance := profitLoss.NetProfit

	privedBalance := 0.0
	for _, v := range equityAccounts {
		amount := struct {
			Sum float64 `sql:"sum"`
		}{}
		err := s.db.Model(&models.TransactionModel{}).
			Where("date <  ?", report.EndDate).
			Select("sum(debit) as sum").
			Joins("JOIN accounts ON accounts.id = transactions.account_id").
			Where("transactions.account_id = ?", v.ID).
			Where("is_opening_balance = ?", false).
			Scan(&amount).Error
		if err != nil {
			return nil, err
		}

		privedBalance += amount.Sum
	}

	capitalChangeBalance := 0.0
	for _, v := range equityAccounts {
		amount := struct {
			Sum float64 `sql:"sum"`
		}{}
		err := s.db.Model(&models.TransactionModel{}).
			Where("date <  ?", report.EndDate).
			Select("sum(credit) as sum").
			Joins("JOIN accounts ON accounts.id = transactions.account_id").
			Where("transactions.account_id = ?", v.ID).
			Where("is_opening_balance = ?", false).
			Scan(&amount).Error
		if err != nil {
			return nil, err
		}

		capitalChangeBalance += amount.Sum
	}

	// amount := struct {
	// 	Sum float64 `sql:"sum"`
	// }{}
	// err := s.db.Model(&models.TransactionModel{}).
	// 	Where("date <  ?", report.StartDate).
	// 	Select("sum(credit-debit) as sum").
	// 	Joins("JOIN accounts ON accounts.id = transactions.account_id").
	// 	Where("accounts.type IN (?)", []models.AccountType{models.EQUITY}).
	// 	Scan(&amount).Error
	// if err != nil {
	// 	return nil, err
	// }

	capitalChange.OpeningBalance = openingBalance
	capitalChange.ProfitLoss = profitLossBalance
	capitalChange.PrivedBalance = -privedBalance
	capitalChange.CapitalChangeBalance = capitalChangeBalance
	capitalChange.EndingBalance = openingBalance + profitLossBalance + capitalChangeBalance - privedBalance
	return &capitalChange, nil
}

func (s *FinanceReportService) GenerateCashFlowReport(report models.GeneralReport) (*models.CashFlowReport, error) {
	cashFlow := models.CashFlowReport{}
	cashFlow.StartDate = report.StartDate
	cashFlow.EndDate = report.EndDate
	cashFlow.Operating = []models.CashflowSubGroup{
		{Name: constants.ACCEPTANCE_FROM_CUSTOMERS, Description: constants.ACCEPTANCE_FROM_CUSTOMERS_VALUE, Amount: 0},
		{Name: constants.OTHER_CURRENT_ASSETS, Description: constants.OTHER_CURRENT_ASSETS_VALUE, Amount: 0},
		{Name: constants.PAYMENT_TO_VENDORS, Description: constants.PAYMENT_TO_VENDORS_VALUE, Amount: 0},
		{Name: constants.CREDIT_CARDS_AND_OTHER_SHORT_TERM_LIABILITIES, Description: constants.CREDIT_CARDS_AND_OTHER_SHORT_TERM_LIABILITIES_VALUE, Amount: 0},
		{Name: constants.OTHER_INCOME, Description: constants.OTHER_INCOME_VALUE, Amount: 0},
		{Name: constants.OPERATIONAL_EXPENSES, Description: constants.OPERATIONAL_EXPENSES_VALUE, Amount: 0},
		{Name: constants.RETURNS_PAYMENT_OF_TAXES, Description: constants.RETURNS_PAYMENT_OF_TAXES_VALUE, Amount: 0},
		{Name: constants.COOPERATIVE_ACCEPTANCE_FROM_MEMBER, Description: constants.COOPERATIVE_ACCEPTANCE_FROM_MEMBER_LABEL, Amount: 0},
		{Name: constants.COOPERATIVE_ACCEPTANCE_FROM_NON_MEMBER, Description: constants.COOPERATIVE_ACCEPTANCE_FROM_NON_MEMBER_LABEL, Amount: 0},
	}
	cashFlow.Investing = []models.CashflowSubGroup{
		{Name: constants.ACQUISITION_SALE_OF_ASSETS, Description: constants.ACQUISITION_SALE_OF_ASSETS_VALUE, Amount: 0},
		{Name: constants.OTHER_INVESTMENT_ACTIVITIES, Description: constants.OTHER_INVESTMENT_ACTIVITIES_VALUE, Amount: 0},
		{Name: constants.INVESTMENT_PARTNERSHIP, Description: constants.INVESTMENT_PARTNERSHIP_VALUE, Amount: 0},
	}
	cashFlow.Financing = []models.CashflowSubGroup{
		{Name: constants.LOAN_PAYMENTS_RECEIPTS, Description: constants.LOAN_PAYMENTS_RECEIPTS_VALUE, Amount: 0},
		{Name: constants.EQUITY_CAPITAL, Description: constants.EQUITY_CAPITAL_VALUE, Amount: 0},
		{Name: constants.COOPERATIVE_PRINCIPAL_SAVING, Description: constants.COOPERATIVE_PRINCIPAL_SAVING_LABEL, Amount: 0},
		{Name: constants.COOPERATIVE_MANDATORY_SAVING, Description: constants.COOPERATIVE_MANDATORY_SAVING_LABEL, Amount: 0},
		{Name: constants.COOPERATIVE_VOLUNTARY_SAVING, Description: constants.COOPERATIVE_VOLUNTARY_SAVING_LABEL, Amount: 0},
	}

	fmt.Println("======================================")
	fmt.Println("OPERATING")
	fmt.Println("======================================")
	operating, totalOperating := s.getCashFlowAmount(cashFlow.Operating)
	cashFlow.Operating = operating
	cashFlow.TotalOperating = totalOperating

	fmt.Println("======================================")
	fmt.Println("INVESTING")
	fmt.Println("======================================")
	investing, totalInvesting := s.getCashFlowAmount(cashFlow.Investing)
	cashFlow.Investing = investing
	cashFlow.TotalInvesting = totalInvesting

	fmt.Println("======================================")
	fmt.Println("FINANCING")
	fmt.Println("======================================")
	financing, totalInvesting := s.getCashFlowAmount(cashFlow.Financing)
	cashFlow.Financing = financing
	cashFlow.TotalFinancing = totalInvesting

	return &cashFlow, nil
}

func (s *FinanceReportService) getCashFlowAmount(groups []models.CashflowSubGroup) ([]models.CashflowSubGroup, float64) {

	total := 0.0
	for i, v := range groups {
		var transactions []models.TransactionModel
		s.db.Model(&transactions).
			Distinct("transRef.id refid, (transRef.debit - transRef.credit) amount, accountRef.name description").
			Joins("JOIN accounts ON accounts.id = transactions.account_id").
			Joins("JOIN transactions transRef ON transRef.id = transactions.transaction_ref_id").
			Joins("JOIN accounts accountRef ON accountRef.id = transRef.account_id").
			Where("accounts.cashflow_sub_group = ?", v.Name).
			Where("accountRef.cashflow_sub_group = ?", "cash_bank").
			Group("refid, transactions.id, accountRef.name").
			Find(&transactions)

		amount := 0.0
		for _, t := range transactions {
			fmt.Printf("[%s] %s %f\n", v.Name, t.Description, t.Amount)
			amount += t.Amount
		}
		v.Amount = amount
		groups[i] = v
		total += amount
	}
	return groups, total
}
