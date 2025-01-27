package models

import (
	"encoding/json"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/AMETORY/ametory-erp-modules/shared/objects"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ShippingModel struct {
	shared.BaseModel
	OfferID              *string                `gorm:"type:char(36)" json:"offer_id,omitempty"`
	OrderID              *string                `gorm:"type:char(36);not null" json:"order_id,omitempty"`    // ID pesanan
	ShippingID           *string                `gorm:"type:char(36);not null" json:"shipping_id,omitempty"` // ID pesanan
	Method               string                 `gorm:"not null" json:"method,omitempty"`                    // Metode pengiriman
	TrackingID           string                 `gorm:"unique" json:"tracking_id,omitempty"`                 // Nomor resi
	Provider             string                 `gorm:"not null" json:"provider,omitempty"`                  // Provider pengiriman
	Status               string                 `gorm:"not null" json:"status,omitempty"`                    // Status terakhir pengiriman
	CourierName          string                 `gorm:"not null" json:"courier_name,omitempty"`
	ServiceType          string                 `gorm:"not null" json:"service_type,omitempty"`
	TrackingData         string                 `gorm:"type:json" json:"tracking_data,omitempty"`
	ShippingData         string                 `gorm:"type:json" json:"-"`
	ShippingDataResponse map[string]interface{} `gorm:"-" json:"shipping_data_response,omitempty"`
	TrackingStatuses     []objects.History      `gorm:"-" json:"tracking_statuses,omitempty"`
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

func (s *ShippingModel) AfterFind(tx *gorm.DB) error {
	if s.ShippingData != "" {
		if err := json.Unmarshal([]byte(s.ShippingData), &s.ShippingDataResponse); err != nil {
			return err
		}
	}
	return nil
}
