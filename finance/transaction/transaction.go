package transaction

import (
	"time"

	"gorm.io/gorm"
)

type TransactionService struct {
	db *gorm.DB
}

func NewTransactionService(db *gorm.DB) *TransactionService {
	return &TransactionService{db: db}
}

func (s *TransactionService) CreateTransaction(transaction *TransactionModel) error {
	return s.db.Create(transaction).Error
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

func (s *TransactionService) GetTransactions(page int, limit int, search string) ([]TransactionModel, error) {
	var accounts []TransactionModel
	query := s.db

	if search != "" {
		query = query.Select("transactions.*, account_source.name as account_source_name, account_destination.name as account_destination_name").
			Joins("LEFT JOIN account as account_source", "account_source.id = transactions.account_source_id").
			Joins("LEFT JOIN account as account_destination", "account_destination.id = transactions.account_destination_id")
		query = query.Where("transactions.description LIKE ? OR transactions.notes LIKE ? OR account_source.name LIKE ? OR account_destination.name LIKE ?", "%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Find(&accounts).Error
	return accounts, err
}
