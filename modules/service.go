package modules

import (
	"gorm.io/gorm"
)

// Service adalah interface yang harus diimplementasikan oleh setiap service
type Service interface {
	Migrate() error // Migrate menjalankan migrasi database
	DB() *gorm.DB   // DB mengembalikan instance database
}
