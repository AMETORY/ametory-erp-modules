package models

import (
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type HealthFacilityModel struct {
	shared.BaseModel
	Type          string             `gorm:"type:varchar(20);not null" json:"type"`
	Name          string             `gorm:"type:varchar(255);not null" json:"name,omitempty"`
	Address       string             `gorm:"type:varchar(255)" json:"address,omitempty"`
	PhoneNumber   string             `gorm:"type:varchar(255)" json:"phone_number,omitempty"`
	Email         string             `gorm:"type:varchar(255)" json:"email,omitempty"`   // TODO: validate email
	Website       string             `gorm:"type:varchar(255)" json:"website,omitempty"` // TODO: validate website
	SubFacilities []SubFacilityModel `gorm:"foreignKey:HealthFacilityID;constraint:OnDelete:CASCADE;" json:"sub_facilities,omitempty"`
}

func (HealthFacilityModel) TableName() string {
	return "health_facilities"
}

func (h HealthFacilityModel) BeforeCreate(tx *gorm.DB) (err error) {
	if h.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return nil
}
