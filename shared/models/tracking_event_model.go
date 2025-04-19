package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TrackingEventModel struct {
	shared.BaseModel
	ShipmentLegID *string           `json:"shipment_leg_id,omitempty"`
	ShipmentLeg   *ShipmentLegModel `gorm:"foreignKey:ShipmentLegID;constraint:OnDelete:CASCADE" json:"shipment_leg,omitempty"`
	Status        string            `json:"status,omitempty" gorm:"type:varchar(20)"`
	LocationName  string            `json:"location_name,omitempty" gorm:"type:varchar(255)"`
	Latitude      float64           `json:"latitude" gorm:"type:decimal(10,8);"`
	Longitude     float64           `json:"longitude" gorm:"type:decimal(11,8);"`
	Timestamp     time.Time         `json:"timestamp,omitempty"`
	Notes         string            `json:"notes,omitempty" gorm:"type:text"`
}

func (t *TrackingEventModel) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return nil
}

func (t *TrackingEventModel) TableName() string {
	return "tracking_events"
}
