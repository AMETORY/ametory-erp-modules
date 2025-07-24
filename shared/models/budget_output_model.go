package models

import (
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BudgetOutputModel is a struct for budget output model
type BudgetOutputModel struct {
	shared.BaseModel
	Name                       string                         `gorm:"type:varchar(255);not null" json:"name"`
	Description                string                         `gorm:"type:text" json:"description"`
	BudgetID                   *string                        `gorm:"type:char(36);index" json:"budget_id"`
	Budget                     *BudgetModel                   `gorm:"foreignKey:BudgetID;constraint:OnDelete:CASCADE;" json:"budget"`
	BudgetStrategicObjectiveID *string                        `gorm:"type:char(36);index" json:"budget_strategic_objective_id"` // use this if explicit target
	BudgetStrategicObjective   *BudgetStrategicObjectiveModel `gorm:"foreignKey:BudgetStrategicObjectiveID;constraint:OnDelete:CASCADE;" json:"budget_strategic_objective"`
	BudgetKPIID                *string                        `gorm:"type:char(36);index" json:"budget_kpi_id"` // use this if explicit target
	BudgetKPI                  *BudgetKPIModel                `gorm:"foreignKey:BudgetKPIID;constraint:OnDelete:CASCADE;" json:"budget_kpi"`
}

// TableName returns the table name for BudgetOutputModel
func (b *BudgetOutputModel) TableName() string {
	return "budget_outputs"
}

// BeforeCreate sets the default ID for BudgetOutputModel
func (b *BudgetOutputModel) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		b.ID = uuid.New().String()
	}
	return nil
}
