package journal

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/finance/account"
	"github.com/AMETORY/ametory-erp-modules/finance/transaction"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type JournalService struct {
	ctx                *context.ERPContext
	db                 *gorm.DB
	accountService     *account.AccountService
	transactionService *transaction.TransactionService
}

func NewJournalService(db *gorm.DB,
	ctx *context.ERPContext,
	accountService *account.AccountService,
	transactionService *transaction.TransactionService,
) *JournalService {
	return &JournalService{
		ctx:                ctx,
		db:                 db,
		accountService:     accountService,
		transactionService: transactionService,
	}
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&models.JournalModel{})
}

func (js *JournalService) CreateJournal(data *models.JournalModel) error {
	return js.db.Create(data).Error
}

func (js *JournalService) GetJournal(id string) (*models.JournalModel, error) {
	var journal models.JournalModel
	err := js.db.Where("id = ?", id).First(&journal).Error
	if err != nil {
		return nil, err
	}
	transactions, err := js.GetTransactions(journal.ID, *js.ctx.Request)
	if err != nil {
		return nil, err
	}
	journal.Transactions = transactions
	var credit, debit float64
	for _, transaction := range transactions {
		credit += transaction.Credit
		debit += transaction.Debit
	}

	// fmt.Println("BALANCE", credit, debit, credit != debit)
	journal.Unbalanced = credit != debit
	return &journal, err
}

func (js *JournalService) UpdateJournal(id string, data *models.JournalModel) error {
	return js.db.Where("id = ?", id).Updates(data).Error
}

func (js *JournalService) DeleteJournal(id string) error {
	var transaction models.TransactionModel
	err := js.db.Where("transaction_ref_id = ? and transaction_ref_type = ?", id, "journal").Delete(&transaction).Error
	if err != nil {
		return err
	}
	return js.db.Where("id = ?", id).Delete(&models.JournalModel{}).Error
}

func (js *JournalService) GetJournals(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := js.db
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}

	stmt = stmt.Model(&models.JournalModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.JournalModel{})
	page.Page = page.Page + 1
	items := page.Items.(*[]models.JournalModel)
	newItems := make([]models.JournalModel, 0)
	for _, item := range *items {
		amount := struct {
			Credit float64 `sql:"credit"`
			Debit  float64 `sql:"debit"`
		}{}
		js.db.Model(&models.TransactionModel{}).Select("sum(credit) as credit, sum(debit) as debit").Where("transaction_ref_id = ?", item.ID).Scan(&amount)
		item.Unbalanced = amount.Credit != amount.Debit
		newItems = append(newItems, item)
	}
	page.Items = &newItems
	return page, nil
}

func (js *JournalService) AddTransaction(journalID string, transaction *models.TransactionModel, amount float64) error {
	transaction.TransactionRefID = &journalID
	transaction.TransactionRefType = "journal"

	return js.transactionService.CreateTransaction(transaction, amount)
}

func (js *JournalService) DeleteTransaction(id string) error {
	return js.transactionService.DeleteTransaction(id)
}

func (js *JournalService) GetTransactions(journalID string, request http.Request) ([]models.TransactionModel, error) {
	var trans []models.TransactionModel
	db := js.db.Preload("Account").Where("transaction_ref_id = ? AND transaction_ref_type = ?", journalID, "journal")
	if request.Header.Get("ID-Company") != "" {
		db = db.Where("company_id = ?", request.Header.Get("ID-Company"))
	}

	switch request.URL.Query().Get("type") {
	case "INCOME":
		db = db.Where("transactions.is_income = ?", true)
	case "EXPENSE":
		db = db.Where("transactions.is_expense = ?", true)
	case "EQUITY":
		db = db.Where("transactions.is_equity = ?", true)
	case "TRANSFER":
		db = db.Where("transactions.is_transfer = ?", true)
	}

	if request.URL.Query().Get("account_id") != "" {
		db = db.Where("transactions.account_id = ?", request.URL.Query().Get("account_id"))
	}

	db = db.Find(&trans)
	err := db.Error
	return trans, err
}
