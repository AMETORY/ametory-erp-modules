package models

import (
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ShippingModel struct {
	shared.BaseModel
	OrderID    string `gorm:"not null"` // ID pesanan
	Method     string `gorm:"not null"` // Metode pengiriman
	TrackingID string `gorm:"unique"`   // Nomor resi
	Provider   string `gorm:"not null"` // Provider pengiriman
	Status     string `gorm:"not null"` // Status terakhir pengiriman
}

func (s *ShippingModel) TableName() string {
	return "shippings"
}

func (s *ShippingModel) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	return nil
}
