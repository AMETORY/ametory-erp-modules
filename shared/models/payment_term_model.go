package models

import (
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PaymentTermModel struct {
	shared.BaseModel
	Name            string   `json:"name"`
	Code            string   `json:"code" gorm:"uniqueIndex"`
	Description     string   `json:"description"`
	Category        string   `json:"category"`
	DueDays         *int     `json:"due_days,omitempty"`
	DiscountAmount  *float64 `json:"discount_amount,omitempty"`
	DiscountDueDays *int     `json:"discount_due_days,omitempty"`
}

func (PaymentTermModel) TableName() string {
	return "payment_terms"
}

func (pt *PaymentTermModel) BeforeCreate(tx *gorm.DB) (err error) {
	if pt.ID == "" {
		pt.ID = uuid.New().String()
	}
	return nil
}
