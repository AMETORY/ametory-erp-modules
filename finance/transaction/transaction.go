package transaction

import (
	"net/http"
	"time"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type TransactionService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewTransactionService(db *gorm.DB, ctx *context.ERPContext) *TransactionService {
	return &TransactionService{db: db, ctx: ctx}
}

func (s *TransactionService) CreateTransaction(transaction *TransactionModel) error {
	if transaction.SourceID != nil {
		transaction.AccountID = transaction.SourceID
		if err := s.db.Create(transaction).Error; err != nil {
			return err
		}
	}
	if transaction.DestinationID != nil {
		transaction.AccountID = transaction.DestinationID
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

func (s *TransactionService) GetTransactionByDate(from, to time.Time, page, limit int) ([]TransactionModel, error) {
	var transactions []TransactionModel
	err := s.db.Where("date BETWEEN ? AND ?", from, to).Offset((page - 1) * limit).Limit(limit).Find(&transactions).Error
	return transactions, err
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
	if search != "" {
		stmt = stmt.Where("accounts.name LIKE ? OR accounts.code LIKE ? OR transactions.code LIKE ? OR transactions.description LIKE ?",
			"%"+search+"%",
			"%"+search+"%",
			"%"+search+"%",
			"%"+search+"%",
		)
	}
	stmt = stmt.Model(&TransactionModel{})
	page := pg.With(stmt).Request(request).Response(&[]TransactionModel{})
	return page, nil
}
