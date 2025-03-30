package models

import (
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TaxModel struct {
	shared.BaseModel
	UserID              *string       `gorm:"size:36" json:"-"`
	User                *UserModel    `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	CompanyID           *string       `gorm:"size:36" json:"-"`
	Company             *CompanyModel `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE" json:"company,omitempty"`
	Name                string        `json:"name"`
	Code                string        `json:"code"`
	Amount              float64       `json:"amount"`
	AccountReceivableID *string       `gorm:"size:36" json:"account_receivable_id"`
	AccountPayableID    *string       `gorm:"size:36" json:"account_payable_id"`
	AccountReceivable   *AccountModel `gorm:"foreignKey:AccountReceivableID;constraint:OnDelete:SET NULL" json:"account_receivable,omitempty"`
	AccountPayable      *AccountModel `gorm:"foreignKey:AccountPayableID;constraint:OnDelete:SET NULL" json:"account_payable,omitempty"`
}

func (t *TaxModel) TableName() string {
	return "taxes"
}

func (t *TaxModel) BeforeCreate(tx *gorm.DB) (err error) {
	if t.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}
