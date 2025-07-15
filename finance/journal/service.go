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

// NewJournalService creates a new instance of JournalService with the provided
// database connection, ERP context, account service, and transaction service.
// It is used to manage journal entries and perform operations related to journals.

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

// Migrate creates the database tables required for the journal service, if they do
// not already exist.
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&models.JournalModel{})
}

// CreateJournal creates a new journal entry based on the provided data.
func (js *JournalService) CreateJournal(data *models.JournalModel) error {
	return js.db.Create(data).Error
}

// GetJournal retrieves a journal entry by its ID along with its associated
// transactions. It calculates the total credit and debit amounts from the
// transactions and determines if the journal is unbalanced. Returns the
// journal model with transactions and unbalanced status, or an error if
// the journal cannot be found or transactions cannot be retrieved.

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

// UpdateJournal updates an existing journal entry based on the provided data.
//
// It takes an ID of the journal entry to be updated and a pointer to a JournalModel
// containing the updated journal information. The function returns an error if the
// update operation fails.
func (js *JournalService) UpdateJournal(id string, data *models.JournalModel) error {
	return js.db.Where("id = ?", id).Updates(data).Error
}

// DeleteJournal deletes a journal entry and its associated transactions.
//
// This function takes the ID of the journal entry to be deleted. It first
// removes all transactions linked to the journal entry by their reference ID
// and type. Then, it deletes the journal entry itself.
//
// Returns an error if any of the delete operations fail; otherwise, nil.
func (js *JournalService) DeleteJournal(id string) error {
	var transaction models.TransactionModel
	err := js.db.Where("transaction_ref_id = ? and transaction_ref_type = ?", id, "journal").Delete(&transaction).Error
	if err != nil {
		return err
	}
	return js.db.Where("id = ?", id).Delete(&models.JournalModel{}).Error
}

// GetJournals retrieves a paginated list of journals from the database.
//
// It takes an HTTP request and a search query string as input. The method uses
// GORM to query the database for journals, applying the search query to the
// journal name and description fields. If the request contains a company ID header,
// the method filters the result by the company ID.
// The function utilizes pagination to manage the result set and applies any
// necessary request modifications using the utils.FixRequest utility.
// The function returns a paginated page of JournalModel and an error if the
// operation fails.
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

// AddTransaction adds a transaction to a journal entry.
//
// It takes the ID of the journal entry, a pointer to a TransactionModel
// containing the transaction information, and the amount of the transaction.
// The function sets the TransactionSecondaryRefID and TransactionSecondaryRefType
// fields of the transaction to the provided journal ID and "journal" respectively,
// and then calls the CreateTransaction method of the TransactionService to
// persist the transaction to the database.
// The function returns an error if the transaction creation fails.
func (js *JournalService) AddTransaction(journalID string, transaction *models.TransactionModel, amount float64) error {
	transaction.TransactionSecondaryRefID = &journalID
	transaction.TransactionSecondaryRefType = "journal"

	return js.transactionService.CreateTransaction(transaction, amount)
}

// DeleteTransaction deletes a transaction by its ID.
//
// The function takes the ID of the transaction as its argument.
// It returns an error if the deletion operation fails.
func (js *JournalService) DeleteTransaction(id string) error {
	return js.transactionService.DeleteTransaction(id)
}

func (js *JournalService) GetTransactions(journalID string, request http.Request) ([]models.TransactionModel, error) {
	var trans []models.TransactionModel
	db := js.db.Preload("Account").Where("(transaction_ref_id = ? AND transaction_ref_type = ?) OR (transaction_secondary_ref_id = ? AND transaction_secondary_ref_type = ?)", journalID, "journal", journalID, "journal")
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
