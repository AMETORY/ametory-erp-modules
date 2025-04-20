package models

import (
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LocationPointModel adalah model database untuk warehouse location
type LocationPointModel struct {
	shared.BaseModel
	Name        string          `gorm:"not null;type:varchar(255)" json:"name,omitempty"`
	WarehouseID *string         `gorm:"size:36" json:"warehouse_id,omitempty"`
	Warehouse   *WarehouseModel `gorm:"foreignKey:WarehouseID;constraint:OnDelete:CASCADE" json:"warehouse,omitempty"`
	Type        string          `gorm:"type:varchar(20)" json:"type,omitempty"` // Warehouse, RegionalHub, Posko, etc
	Address     string          `gorm:"type:varchar(255)" json:"address,omitempty"`
	Latitude    float64         `json:"latitude" gorm:"type:decimal(10,8);"`
	Longitude   float64         `json:"longitude" gorm:"type:decimal(11,8);"`
}

func (m *LocationPointModel) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}

func (m *LocationPointModel) TableName() string {
	return "location_points"
}
