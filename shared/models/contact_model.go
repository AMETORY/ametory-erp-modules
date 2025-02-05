package models

import (
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ContactModel adalah model database untuk contact
type ContactModel struct {
	shared.BaseModel
	Name                  string        `gorm:"not null" json:"name"`
	Email                 string        `json:"email"`
	Code                  string        `json:"code"`
	Phone                 *string       `json:"phone"`
	Address               string        `json:"address"`
	ContactPerson         string        `json:"contact_person"`
	ContactPersonPosition string        `json:"contact_person_position"`
	IsCustomer            bool          `gorm:"default:false" json:"is_customer"` // Flag untuk customer
	IsVendor              bool          `gorm:"default:false" json:"is_vendor"`   // Flag untuk vendor
	IsSupplier            bool          `gorm:"default:false" json:"is_supplier"` // Flag untuk supplier
	UserID                *string       `json:"user_id" gorm:"user_id"`
	User                  *UserModel    `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	CompanyID             *string       `json:"company_id" gorm:"company_id"`
	Company               *CompanyModel `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE"`
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
