package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DistributionEventModel adalah model database untuk distribution event
type DistributionEventModel struct {
	shared.BaseModel
	Name        string          `gorm:"not null" json:"name,omitempty"`
	Description string          `json:"description,omitempty"`
	StartDate   time.Time       `json:"start_date,omitempty"`
	EndDate     *time.Time      `json:"end_date,omitempty"`
	Shipments   []ShipmentModel `gorm:"foreignKey:DistributionEventID" json:"shipments,omitempty"`
}

func (DistributionEventModel) TableName() string {
	return "distribution_events"
}

func (d *DistributionEventModel) BeforeCreate(tx *gorm.DB) (err error) {
	// Add your custom logic here
	if d.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return nil
}

type DistributionEventReport struct {
	shared.BaseModel
	DistributionEvent   DistributionEventModel `gorm:"foreignKey:DistributionEventID" json:"distribution_event,omitempty"`
	DistributionEventID string                 `gorm:"size:36" json:"distribution_event_id,omitempty"`
	TotalShipments      int                    `json:"total_shipments,omitempty"`
	TotalDestinations   int                    `json:"total_destinations,omitempty"`
	TotalItems          int                    `json:"total_items,omitempty"`
	LostItems           int                    `json:"lost_items,omitempty"`
	DamagedItems        int                    `json:"damaged_items,omitempty"`
	DelayedShipments    int                    `json:"delayed_shipments,omitempty"`
	FinishedShipments   int                    `json:"finished_shipments,omitempty"`
	ProcessingShipments int                    `json:"processing_shipments,omitempty"`
	ReadyToShip         int                    `json:"ready_to_ship,omitempty"`
	FeedbackCount       int                    `json:"feedback_count,omitempty"`
}

func (DistributionEventReport) TableName() string {
	return "distribution_event_reports"
}

func (d *DistributionEventReport) BeforeCreate(tx *gorm.DB) (err error) {
	if d.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return nil
}
