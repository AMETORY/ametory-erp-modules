package modules

import (
	"fmt"

	"github.com/AMETORY/ametory-erp-modules/auth"
	"github.com/AMETORY/ametory-erp-modules/company"
	"github.com/AMETORY/ametory-erp-modules/contact"
	"github.com/AMETORY/ametory-erp-modules/finance"
	"gorm.io/gorm"
)

// Service adalah interface yang harus diimplementasikan oleh setiap service
type Service interface {
	Migrate() error // Migrate menjalankan migrasi database
	DB() *gorm.DB   // DB mengembalikan instance database
}

func RegisterService(serviceName string, db *gorm.DB, skipMigrate bool) (Service, error) {
	switch serviceName {
	case "auth":
		return auth.NewAuthService(db), nil
	case "company":
		return company.NewCompanyService(db, skipMigrate), nil
	case "finance":
		return finance.NewFinanceService(db, skipMigrate), nil
	case "contact":
		return contact.NewContactService(db, skipMigrate), nil

	default:
		return nil, fmt.Errorf("service '%s' tidak ditemukan", serviceName)
	}
}
