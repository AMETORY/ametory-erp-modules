package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CompanyModel struct {
	shared.BaseModel
	Name                  string           `json:"name"`
	Logo                  string           `json:"logo"`
	Cover                 string           `json:"cover"`
	LegalEntity           string           `json:"legal_entity"`
	Email                 string           `json:"email"`
	Phone                 string           `json:"phone"`
	Fax                   string           `json:"fax"`
	Address               string           `json:"address"`
	ContactPerson         string           `json:"contact_person"`
	ContactPersonPosition string           `json:"contact_person_position"`
	TaxPayerNumber        string           `json:"tax_payer_number,omitempty"`
	UserID                *string          `json:"user_id,omitempty" gorm:"constraint:OnDelete:CASCADE;"`
	User                  *UserModel       `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;"`
	Status                string           `json:"status" gorm:"type:VARCHAR(20);DEFAULT:'ACTIVE'"`
	EmployeeActiveCount   int64            `json:"employee_active_count,omitempty"`
	EmployeeResignCount   int64            `json:"employee_resign_count,omitempty"`
	EmployeePendingCount  int64            `json:"employee_pending_count,omitempty"`
	Merchants             []MerchantModel  `json:"merchants,omitempty" gorm:"-"`
	Warehouses            []WarehouseModel `json:"warehouses,omitempty" gorm:"-"`
	EmailVerifiedAt       *time.Time       `gorm:"index" json:"email_verified_at,omitempty"`
	Users                 []UserModel      `gorm:"many2many:user_companies;" json:"users,omitempty"`
}

func (CompanyModel) TableName() string {
	return "companies"
}

func (c *CompanyModel) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return nil
}
