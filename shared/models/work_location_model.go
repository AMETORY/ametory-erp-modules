package models

import "github.com/AMETORY/ametory-erp-modules/shared"

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
