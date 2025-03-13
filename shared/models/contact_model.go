package models

import (
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ContactModel adalah model database untuk contact
type ContactModel struct {
	shared.BaseModel
	Name                  string        `gorm:"not null" json:"name,omitempty"`
	Email                 string        `json:"email,omitempty"`
	Code                  string        `json:"code,omitempty"`
	Phone                 *string       `json:"phone,omitempty"`
	Address               string        `json:"address,omitempty"`
	ContactPerson         string        `json:"contact_person,omitempty"`
	ContactPersonPosition string        `json:"contact_person_position,omitempty"`
	IsCustomer            bool          `gorm:"default:false" json:"is_customer,omitempty"` // Flag untuk customer
	IsVendor              bool          `gorm:"default:false" json:"is_vendor,omitempty"`   // Flag untuk vendor
	IsSupplier            bool          `gorm:"default:false" json:"is_supplier,omitempty"` // Flag untuk supplier
	UserID                *string       `json:"user_id,omitempty" gorm:"user_id"`
	User                  *UserModel    `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	CompanyID             *string       `json:"company_id,omitempty" gorm:"company_id"`
	Company               *CompanyModel `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE" json:"company,omitempty"`
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
