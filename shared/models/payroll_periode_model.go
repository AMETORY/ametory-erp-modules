package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PayRollPeriodeModel struct {
	shared.BaseModel
	CompanyID string         `json:"company_id,omitempty"`
	Company   CompanyModel   `gorm:"foreignKey:CompanyID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"company,omitempty"`
	StartDate time.Time      `json:"start_date,omitempty"`
	EndDate   time.Time      `json:"end_date,omitempty"`
	Notes     string         `json:"notes,omitempty"`
	IsDefault bool           `json:"is_default,omitempty"`
	Status    string         `json:"status,omitempty" gorm:"default:'DRAFT'"`
	Period    string         `json:"period,omitempty"`
	PayRolls  []PayRollModel `gorm:"foreignKey:PayRollPeriodeID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"pay_rolls,omitempty"`
}

func (PayRollPeriodeModel) TableName() string {
	return "pay_roll_periods"
}

func (p *PayRollPeriodeModel) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}
