package models

import (
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PaymentModel adalah model database untuk payment
type PaymentModel struct {
	shared.BaseModel
	Code                string      `gorm:"type:varchar(50);not null;uniqueIndex:idx_code,unique" json:"code"`
	Name                string      `gorm:"type:varchar(255);not null" json:"name"`
	Email               string      `gorm:"type:varchar(255);not null" json:"email"`
	Phone               string      `gorm:"type:varchar(50);not null" json:"phone"`
	Total               float64     `gorm:"type:decimal(10,2);not null" json:"total"`
	PaymentProvider     string      `gorm:"type:varchar(255);not null" json:"payment_provider"`
	PaymentLink         string      `gorm:"type:varchar(255);not null" json:"payment_link"`
	PaymentData         string      `gorm:"type:json" json:"-"`
	PaymentDataResponse interface{} `gorm:"-" json:"payment_data_response"`
	RefID               string      `gorm:"type:varchar(255);not null" json:"ref_id"`
	Status              string      `gorm:"type:varchar(50);default:PENDING;not null" json:"status"`
}

func (s *PaymentModel) TableName() string {
	return "payments"
}

func (pm *PaymentModel) BeforeCreate(tx *gorm.DB) (err error) {
	if pm.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}
