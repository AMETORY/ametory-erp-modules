package models

import (
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WorkLocationModel struct {
	shared.BaseModel
	Name        string        `json:"name"`
	Address     string        `json:"address"`
	Description string        `json:"description"`
	Latitude    *float64      `json:"latitude,omitempty" gorm:"type:DECIMAL(10,8)"`
	Longitude   *float64      `json:"longitude,omitempty" gorm:"type:DECIMAL(11,8)"`
	CompanyID   *string       `json:"company_id"`
	Company     *CompanyModel `gorm:"foreignKey:CompanyID"`
	IsActive    bool          `json:"is_active"`
}

func (WorkLocationModel) TableName() string {
	return "work_locations"
}

func (m *WorkLocationModel) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	return nil
}
