package account

import "gorm.io/gorm"

type AccountService struct {
	db *gorm.DB
}

func NewAccountService(db *gorm.DB) *AccountService {
	return &AccountService{db: db}
}

func (s *AccountService) CreateAccount(name string) error {
	// Implementasi logika bisnis untuk membuat akun
	return nil
}

func (s *AccountService) GetAccountByID(id uint) (interface{}, error) {
	// Implementasi logika bisnis untuk mendapatkan akun
	return nil, nil
}
