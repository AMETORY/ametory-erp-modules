package models

import (
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/AMETORY/ametory-erp-modules/shared/objects"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ShippingModel struct {
	shared.BaseModel
	OrderID          *string                  `gorm:"not null" json:"order_id,omitempty"`  // ID pesanan
	Method           string                   `gorm:"not null" json:"method,omitempty"`    // Metode pengiriman
	TrackingID       string                   `gorm:"unique" json:"tracking_id,omitempty"` // Nomor resi
	Provider         string                   `gorm:"not null" json:"provider,omitempty"`  // Provider pengiriman
	Status           string                   `gorm:"not null" json:"status,omitempty"`    // Status terakhir pengiriman
	TrackingData     string                   `gorm:"type:json" json:"tracking_data,omitempty"`
	TrackingStatuses []objects.TrackingStatus `gorm:"-" json:"tracking_statuses,omitempty"`
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
