package models

import (
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ContactModel adalah model database untuk contact
type ContactModel struct {
	shared.BaseModel
	Name                   string         `gorm:"not null" json:"name,omitempty"`
	Email                  string         `json:"email,omitempty"`
	Code                   string         `json:"code,omitempty"`
	Phone                  *string        `json:"phone,omitempty"`
	Address                string         `json:"address,omitempty"`
	ContactPerson          string         `json:"contact_person,omitempty"`
	ContactPersonPosition  string         `json:"contact_person_position,omitempty"`
	IsCustomer             bool           `gorm:"default:false" json:"is_customer,omitempty"` // Flag untuk customer
	IsVendor               bool           `gorm:"default:false" json:"is_vendor,omitempty"`   // Flag untuk vendor
	IsSupplier             bool           `gorm:"default:false" json:"is_supplier,omitempty"` // Flag untuk supplier
	UserID                 *string        `json:"user_id,omitempty" gorm:"user_id"`
	User                   *UserModel     `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	CompanyID              *string        `json:"company_id,omitempty" gorm:"company_id"`
	Company                *CompanyModel  `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE" json:"company,omitempty"`
	Tags                   []TagModel     `gorm:"many2many:contact_tags;constraint:OnDelete:CASCADE;" json:"tags,omitempty"`
	Count                  int            `gorm:"-" json:"count" sql:"count"`
	Color                  string         `json:"color" gorm:"-" sql:"color"`
	IsCompleted            bool           `json:"is_completed" gorm:"-" sql:"is_completed"`
	IsSuccess              bool           `json:"is_success" gorm:"-" sql:"is_success"`
	Data                   any            `json:"data" gorm:"-"`
	Products               []ProductModel `gorm:"many2many:contact_products;constraint:OnDelete:CASCADE;" json:"products,omitempty"`
	ReceivablesLimit       float64        `gorm:"default:0" json:"receivables_limit"`
	DebtLimit              float64        `gorm:"default:0" json:"debt_limit"`
	ReceivablesLimitRemain float64        `gorm:"-" json:"receivables_limit_remain"`
	DebtLimitRemain        float64        `gorm:"-" json:"debt_limit_remain"`
	TotalDebt              float64        `gorm:"-" json:"total_debt"`
	TotalReceivable        float64        `gorm:"-" json:"total_receivable"`
	TelegramID             *string        `json:"telegram_id"`
	ConnectionType         *string        `json:"connection_type" gorm:"default:whatsapp"`
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

type CountByTag struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
	Count int    `json:"count"`
}
