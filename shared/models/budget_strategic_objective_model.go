package models

import (
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BudgetStrategicObjectiveModel is a struct for budget strategic objective model
type BudgetStrategicObjectiveModel struct {
	shared.BaseModel
	Name        string       `gorm:"type:varchar(255);not null" json:"name"`
	Description string       `gorm:"type:text" json:"description"`
	BudgetID    *string      `gorm:"type:char(36);index" json:"budget_id"`
	Budget      *BudgetModel `gorm:"foreignKey:BudgetID;constraint:OnDelete:CASCADE;" json:"budget"`
}

// TableName returns the table name for BudgetStrategicObjectiveModel
func (b *BudgetStrategicObjectiveModel) TableName() string {
	return "budget_strategic_objectives"
}

// BeforeCreate sets the default ID for BudgetStrategicObjectiveModel
func (b *BudgetStrategicObjectiveModel) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		b.ID = uuid.New().String()
	}
	return nil
}
