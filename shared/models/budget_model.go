package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BudgetStatus string

const (
	BudgetStatusDraft     BudgetStatus = "draft"
	BudgetStatusSubmitted BudgetStatus = "submitted"
	BudgetStatusApproved  BudgetStatus = "approved"
	BudgetStatusRejected  BudgetStatus = "rejected"
)

// BudgetModel is a struct for budget model
type BudgetModel struct {
	shared.BaseModel
	Name                string                          `gorm:"type:varchar(255);not null" json:"name"`
	Description         string                          `gorm:"type:text" json:"description"`
	TotalBudget         float64                         `json:"total_budget"`
	Status              BudgetStatus                    `json:"status"`
	StartDate           *time.Time                      `json:"start_date"`
	EndDate             *time.Time                      `json:"end_date"`
	KPIs                []BudgetKPIModel                `gorm:"foreignKey:BudgetID;constraint:OnDelete:CASCADE" json:"kpis"`
	StrategicObjectives []BudgetStrategicObjectiveModel `gorm:"foreignKey:BudgetID;constraint:OnDelete:CASCADE" json:"strategic_objectives"`
}

// TableName returns the table name for BudgetModel
func (b *BudgetModel) TableName() string {
	return "budgets"
}

// BeforeCreate sets the default ID for BudgetModel
func (b *BudgetModel) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		b.ID = uuid.New().String()
	}
	return nil
}
