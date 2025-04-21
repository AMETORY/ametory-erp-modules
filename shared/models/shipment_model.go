package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ShipmentModel struct {
	shared.BaseModel
	DistributionEvent    *DistributionEventModel `gorm:"foreignKey:DistributionEventID;constraint:OnDelete:CASCADE" json:"distribution_event,omitempty"`
	DistributionEventID  *string                 `gorm:"size:36" json:"distribution_event_id,omitempty"`
	Code                 string                  `gorm:"unique;not null;type:varchar(255)" json:"code,omitempty"`  // kode pengiriman, bisa di-scan
	Status               string                  `gorm:"type:varchar(20);default:PENDING" json:"status,omitempty"` // PENDING, DELIVERED, IN_DELIVERY, CANCELLED
	ShipmentLegs         []ShipmentLegModel      `gorm:"foreignKey:ShipmentID" json:"shipment_legs,omitempty"`
	Items                []ShipmentItem          `gorm:"foreignKey:ShipmentID" json:"items,omitempty"`
	Feedbacks            []ShipmentFeedback      `gorm:"foreignKey:ShipmentID" json:"feedbacks,omitempty"`
	IncidentEvents       []IncidentEventModel    `gorm:"foreignKey:ShipmentID" json:"incident_events,omitempty"`
	CurrentShipmentLegID *string                 `gorm:"size:36" json:"current_shipment_leg_id,omitempty"`
	CurrentShipmentLeg   *ShipmentLegModel       `gorm:"-" json:"current_shipment_leg,omitempty"`
	ShipmentDate         time.Time               `json:"shipment_date,omitempty"`
	ExpectedFinishAt     *time.Time              `json:"expected_finish_at,omitempty"`
	IsDelayed            bool                    `json:"is_delayed,omitempty"`
	Notes                string                  `gorm:"type:text" json:"notes,omitempty"`
	FromLocation         *LocationPointModel     `gorm:"foreignKey:FromLocationID;constraint:OnDelete:CASCADE" json:"from_location,omitempty"`
	FromLocationID       *string                 `gorm:"size:36" json:"from_location_id,omitempty"`
	ToLocation           *LocationPointModel     `gorm:"foreignKey:ToLocationID;constraint:OnDelete:CASCADE" json:"to_location,omitempty"`
	ToLocationID         *string                 `gorm:"size:36" json:"to_location_id,omitempty"`
}

func (ShipmentModel) TableName() string {
	return "shipments"
}

func (s *ShipmentModel) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	// Add logic to execute before creating a ShipmentModel record, if needed
	return nil
}

func (s *ShipmentModel) AfterFind(tx *gorm.DB) (err error) {
	if s.CurrentShipmentLegID != nil {
		var shipmentLeg ShipmentLegModel
		tx.Where("id = ?", *s.CurrentShipmentLegID).First(&shipmentLeg)
		s.CurrentShipmentLeg = &shipmentLeg
	}
	return nil
}

type ShipmentItem struct {
	shared.BaseModel
	Shipment   *ShipmentModel `gorm:"foreignKey:ShipmentID;constraint:OnDelete:CASCADE" json:"shipment,omitempty"`
	ShipmentID *string        `gorm:"size:36" json:"shipment_id,omitempty"`
	ItemName   string         `gorm:"type:varchar(255)" json:"item_name,omitempty"`
	Quantity   float64        `json:"quantity,omitempty"`
	Product    *ProductModel  `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE" json:"product,omitempty"`
	ProductID  *string        `gorm:"size:36" json:"product_id,omitempty"`
	Unit       *UnitModel     `gorm:"foreignKey:UnitID;constraint:OnDelete:CASCADE" json:"unit,omitempty"`
	UnitID     *string        `gorm:"size:36" json:"unit_id,omitempty"`
}

func (s *ShipmentItem) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return nil
}

type ShipmentLegModel struct {
	shared.BaseModel
	Shipment       *ShipmentModel       `gorm:"foreignKey:ShipmentID;constraint:OnDelete:CASCADE" json:"shipment,omitempty"`
	ShipmentID     *string              `gorm:"size:36" json:"shipment_id,omitempty"`
	FromLocation   *LocationPointModel  `gorm:"foreignKey:FromLocationID;constraint:OnDelete:CASCADE" json:"from_location,omitempty"`
	FromLocationID *string              `gorm:"size:36" json:"from_location_id,omitempty"`
	ToLocation     *LocationPointModel  `gorm:"foreignKey:ToLocationID;constraint:OnDelete:CASCADE" json:"to_location,omitempty"`
	ToLocationID   *string              `gorm:"size:36" json:"to_location_id,omitempty"`
	TransportMode  string               `gorm:"type:varchar(50)" json:"transport_mode,omitempty"` // Truck, Boat, Air, Manual, etc
	NumberPlate    string               `gorm:"type:varchar(20)" json:"number_plate,omitempty"`
	DriverName     string               `gorm:"type:varchar(100)" json:"driver_name,omitempty"`
	VehicleInfo    string               `gorm:"type:varchar(255)" json:"vehicle_info,omitempty"`
	Status         string               `gorm:"type:varchar(20)" json:"status,omitempty"` // Pending, In Transit, Completed
	DepartedAt     *time.Time           `json:"departed_at,omitempty"`
	ArrivedAt      *time.Time           `json:"arrived_at,omitempty"`
	TrackingEvents []TrackingEventModel `gorm:"foreignKey:ShipmentLegID;constraint:OnDelete:CASCADE" json:"tracking_events,omitempty"`
	ShippedByID    *string              `gorm:"size:36" json:"shipped_by_id,omitempty"`
	ShippedBy      *UserModel           `gorm:"foreignKey:ShippedByID;constraint:OnDelete:SET NULL" json:"shipped_by,omitempty"`
	ArrivedByID    *string              `gorm:"size:36" json:"arrived_by_id,omitempty"`
	ArrivedBy      *UserModel           `gorm:"foreignKey:ArrivedByID;constraint:OnDelete:SET NULL" json:"arrived_by,omitempty"`
}

func (ShipmentLegModel) TableName() string {
	return "shipment_legs"
}

func (s *ShipmentLegModel) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	// Add logic to execute before creating a ShipmentLegModel record, if needed
	return nil
}

type ShipmentFeedback struct {
	shared.BaseModel
	Shipment     *ShipmentModel `gorm:"foreignKey:ShipmentID;constraint:OnDelete:CASCADE" json:"shipment,omitempty"`
	ShipmentID   *string        `gorm:"size:36" json:"shipment_id,omitempty"`
	GiverID      *string        `gorm:"size:36" json:"giver_id,omitempty"` // optional, jika feedback dari user terdaftar
	Giver        *UserModel     `gorm:"foreignKey:GiverID;constraint:OnDelete:SET NULL" json:"giver,omitempty"`
	GiverName    string         `gorm:"type:varchar(100)" json:"giver_name,omitempty"`    // nama bebas dari pemberi feedback
	GiverContact *string        `gorm:"type:varchar(100)" json:"giver_contact,omitempty"` // opsional, email/telp jika ada
	GiverRole    *string        `gorm:"type:varchar(50)" json:"giver_role,omitempty"`     // "Receiver", "Volunteer", "LocalAuthority", etc.
	Rating       int            `gorm:"type:smallint" json:"rating,omitempty"`            // 1–5
	Comment      string         `gorm:"type:text" json:"comment,omitempty"`
}

func (s *ShipmentFeedback) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return nil
}
