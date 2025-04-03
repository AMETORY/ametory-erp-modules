package models

import (
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UnitModel struct {
	shared.BaseModel
	Name        string        `gorm:"type:varchar(255);not null" json:"name"`
	Code        string        `gorm:"type:varchar(255);not null" json:"code"`
	Description string        `json:"description"`
	CompanyID   *string       `json:"company_id,omitempty"`
	Company     *CompanyModel `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE" json:"company,omitempty"`
	IsDefault   bool          `json:"is_default,omitempty" gorm:"-"`
	Value       float64       `gorm:"-" json:"value,omitempty"`
}

func (UnitModel) TableName() string {
	return "units"
}

func (u *UnitModel) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	// Add custom logic before creating a ProductUnits entry if needed
	return nil
}

type ProductUnits struct {
	IsDefault bool    `json:"is_default,omitempty"`
	Value     float64 `gorm:"default:1"`
}

type ProductUnitData struct {
	ProductModelID *string ` json:"product_model_id,omitempty"`
	UnitModelID    *string ` json:"unit_model_id,omitempty"`
	IsDefault      bool    `json:"is_default,omitempty"`
	Value          float64 `gorm:"default:1"`
}

func (ProductUnitData) TableName() string {
	return "product_units"
}
