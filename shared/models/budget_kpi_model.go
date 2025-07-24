package models

import (
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AchievementType string
type RelatedEntityType string

const (
	AchievementTypeHigherIsBetter AchievementType = "HigherIsBetter"
	AchievementTypeLowerIsBetter  AchievementType = "LowerIsBetter"
	AchievementTypeTargetIsBest   AchievementType = "TargetIsBest"
)

const (
	RelatedEntityTypeBudget             = "Budget"
	RelatedEntityTypeStrategicObjective = "StrategicObjective"
	RelatedEntityTypeOutput             = "Output" // Perubahan dari Program ke Output
	RelatedEntityTypeDivision           = "Division"
	RelatedEntityTypeOverall            = "Overall"
)

// BudgetKPIModel is a struct for budget KPI model
type BudgetKPIModel struct {
	shared.BaseModel
	Name                       string                         `gorm:"type:varchar(255);not null" json:"name"`
	Description                string                         `gorm:"type:text" json:"description"`
	UnitOfMeasure              string                         `json:"unit_of_measure"`
	RelatedEntityType          *RelatedEntityType             `json:"related_entity_type"` // use this if implicit target and use with related entity id
	RelatedEntityID            *string                        `json:"related_entity_id"`
	Target                     float64                        `json:"target"`
	AchievementType            AchievementType                `json:"achievement_type"`
	Frequency                  string                         `json:"frequency,omitempty"`
	BudgetID                   *string                        `gorm:"type:char(36);index" json:"budget_id"`
	Budget                     *BudgetModel                   `gorm:"foreignKey:BudgetID;constraint:OnDelete:CASCADE;" json:"budget"`
	BudgetStrategicObjectiveID *string                        `gorm:"type:char(36);index" json:"budget_strategic_objective_id"` // use this if explicit target
	BudgetStrategicObjective   *BudgetStrategicObjectiveModel `gorm:"foreignKey:BudgetStrategicObjectiveID;constraint:OnDelete:CASCADE;" json:"budget_strategic_objective"`
}

// TableName returns the table name for BudgetKPIModel
func (b *BudgetKPIModel) TableName() string {
	return "budget_kpis"
}

// BeforeCreate sets the default ID for BudgetKPIModel
func (b *BudgetKPIModel) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		b.ID = uuid.New().String()
	}
	return nil
}
