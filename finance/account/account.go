package account

import "gorm.io/gorm"

type AccountService struct {
	db *gorm.DB
}

func NewAccountService(db *gorm.DB) *AccountService {
	return &AccountService{db: db}
}

func (s *AccountService) CreateAccount(data *AccountModel) error {
	return s.db.Create(data).Error
}

func (s *AccountService) UpdateAccount(id string, data *AccountModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

func (s *AccountService) DeleteAccount(id string) error {
	return s.db.Where("id = ?", id).Delete(&AccountModel{}).Error
}

func (s *AccountService) GetAccountByID(id string) (*AccountModel, error) {
	var account AccountModel
	err := s.db.Where("id = ?", id).First(&account).Error
	return &account, err
}

func (s *AccountService) GetAccountByCode(code string) (*AccountModel, error) {
	var account AccountModel
	err := s.db.Where("code = ?", code).First(&account).Error
	return &account, err
}

func (s *AccountService) GetAccounts(page int, limit int, search string) ([]AccountModel, error) {
	var accounts []AccountModel
	query := s.db

	if search != "" {
		query = query.Where("name LIKE ? OR code LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Find(&accounts).Error
	return accounts, err
}
