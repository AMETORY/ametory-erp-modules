package contact

import (
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ContactModel adalah model database untuk contact
type ContactModel struct {
	utils.BaseModel
	Name                  string `gorm:"not null"`
	Email                 string
	Code                  string
	Phone                 string
	Address               string
	ContactPerson         string
	ContactPersonPosition string
	IsCustomer            bool `gorm:"default:false"` // Flag untuk customer
	IsVendor              bool `gorm:"default:false"` // Flag untuk vendor
	IsSupplier            bool `gorm:"default:false"` // Flag untuk supplier
}

func (ContactModel) TableName() string {
	return "contacts"
}

func (u *ContactModel) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}

// Migrate menjalankan migrasi database untuk model contact
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&ContactModel{})
}
