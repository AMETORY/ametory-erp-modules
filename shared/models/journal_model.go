package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type JournalModel struct {
	shared.BaseModel
	UserID           *string            `gorm:"size:30" json:"-" `
	User             *UserModel         `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:UserID" json:"user,omitempty"`
	CompanyID        *string            `gorm:"size:30" json:"-" `
	Company          *CompanyModel      `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:CompanyID" json:"company,omitempty"`
	Description      string             `json:"description"`
	Date             *time.Time         `json:"date"`
	Transactions     []TransactionModel `json:"transactions" gorm:"-"`
	EmployeeID       *string            `gorm:"size:30" json:"employee_id,omitempty"`
	IsOpeningBalance bool               `json:"is_opening_balance"`
	Unbalanced       bool               `json:"unbalanced" gorm:"-"`
}

func (JournalModel) TableName() string {
	return "journals"
}

func (j *JournalModel) BeforeCreate(tx *gorm.DB) error {
	if j.ID == "" {
		j.ID = uuid.New().String()
	}
	return nil
}
