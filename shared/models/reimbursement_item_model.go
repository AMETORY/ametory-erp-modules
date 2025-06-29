package models

import (
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReimbursementItemModel struct {
	shared.BaseModel
	Amount          float64             `json:"amount" form:"amount"`
	Notes           string              `json:"notes" form:"notes"`
	ReimbursementID *string             `json:"reimbursement_id"`
	Reimbursement   *ReimbursementModel `gorm:"foreignKey:ReimbursementID" json:"-"`
	Attachments     []FileModel         `json:"attachments" gorm:"-"`
	CompanyID       *string             `json:"company_id" gorm:"not null"`
	Company         *CompanyModel       `gorm:"foreignKey:CompanyID"`
}

func (ReimbursementItemModel) TableName() string {
	return "reimbursement_items"
}

func (r *ReimbursementItemModel) BeforeCreate(tx *gorm.DB) (err error) {

	if r.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}
