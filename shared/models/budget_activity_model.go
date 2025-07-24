package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BudgetActivityModel is a struct for budget activity model
type BudgetActivityModel struct {
	shared.BaseModel
	Name                 string                `gorm:"type:varchar(255);not null" json:"name"`
	Description          string                `gorm:"type:text" json:"description"`
	BudgetID             *string               `gorm:"type:char(36);index" json:"budget_id"`
	Budget               *BudgetModel          `gorm:"foreignKey:BudgetID;constraint:OnDelete:CASCADE;" json:"budget"`
	BudgetComponentID    *string               `gorm:"type:char(36);index" json:"budget_component_id"`
	BudgetComponentModel *BudgetComponentModel `gorm:"foreignKey:BudgetComponentID;constraint:OnDelete:CASCADE;" json:"budget_component"`
	PlannedBudget        float64               `json:"planned_budget"` // Anggaran yang direncanakan untuk aktivitas ini
	StartDate            *time.Time            `json:"start_date"`
	EndDate              *time.Time            `json:"end_date"`
}

// TableName returns the table name for BudgetActivityModel
func (b *BudgetActivityModel) TableName() string {
	return "budget_activities"
}

// BeforeCreate sets the default ID for BudgetActivityModel
func (b *BudgetActivityModel) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		b.ID = uuid.New().String()
	}
	return nil
}
