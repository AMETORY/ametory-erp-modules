package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IncidentEventModel struct {
	shared.BaseModel
	ShipmentID          *string                 `gorm:"size:36" json:"shipment_id,omitempty"` // optional, if the incident occurs during shipment
	Shipment            *ShipmentModel          `gorm:"foreignKey:ShipmentID;constraint:OnDelete:CASCADE" json:"shipment,omitempty"`
	ShipmentLegID       *string                 `gorm:"size:36" json:"shipment_leg_id,omitempty"`
	ShipmentLeg         *ShipmentLegModel       `gorm:"foreignKey:ShipmentLegID;constraint:OnDelete:CASCADE" json:"shipment_leg,omitempty"`
	WarehouseLocationID *string                 `gorm:"size:36" json:"warehouse_location_id,omitempty"`
	WarehouseLocation   *WarehouseLocationModel `gorm:"foreignKey:WarehouseLocationID;constraint:OnDelete:CASCADE" json:"warehouse_location,omitempty"`
	EventType           string                  `gorm:"type:varchar(50)" json:"event_type,omitempty"` // e.g., "Lost", "Damaged", "Spoiled"
	Description         string                  `gorm:"type:text" json:"description,omitempty"`
	OccurredAt          time.Time               `json:"occurred_at,omitempty"`
	ReportedByID        *string                 `gorm:"size:36" json:"reported_by_id,omitempty"`
	ReportedBy          *UserModel              `gorm:"foreignKey:ReportedByID;constraint:OnDelete:CASCADE" json:"reported_by,omitempty"`
	Items               []IncidentItem          `gorm:"foreignKey:IncidentID" json:"items,omitempty"`
}

func (ie *IncidentEventModel) BeforeCreate(tx *gorm.DB) (err error) {
	if ie.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}

func (ie *IncidentEventModel) TableName() string {
	return "incident_events"
}

type IncidentItem struct {
	shared.BaseModel
	IncidentID  *string             `gorm:"size:36" json:"incident_id,omitempty"`
	Incident    *IncidentEventModel `gorm:"foreignKey:IncidentID;constraint:OnDelete:CASCADE" json:"incident,omitempty"`
	ProductID   *string             `gorm:"size:36" json:"product_id,omitempty"`
	Product     *ProductModel       `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE" json:"product,omitempty"`
	QtyAffected float64             `json:"qty_affected,omitempty"`
	UnitID      *string             `gorm:"size:36" json:"unit_id,omitempty"`
	Unit        *UnitModel          `gorm:"foreignKey:UnitID;constraint:OnDelete:CASCADE" json:"unit,omitempty"`
	Notes       string              `gorm:"type:varchar(255)" json:"notes,omitempty"` // misalnya "kemasan rusak", "hilang saat transit"
}

func (ii *IncidentItem) BeforeCreate(tx *gorm.DB) (err error) {
	if ii.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}
