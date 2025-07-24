package models

import (
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BudgetComponentModel is a struct for budget component model
type BudgetComponentModel struct {
	shared.BaseModel
	Name           string             `gorm:"type:varchar(255);not null" json:"name"`
	Description    string             `gorm:"type:text" json:"description"`
	BudgetID       *string            `gorm:"type:char(36);index" json:"budget_id"`
	Budget         *BudgetModel       `gorm:"foreignKey:BudgetID;constraint:OnDelete:CASCADE;" json:"budget"`
	BudgetOutputID *string            `gorm:"type:char(36);index" json:"budget_output_id"`
	BudgetOutput   *BudgetOutputModel `gorm:"foreignKey:BudgetOutputID;constraint:OnDelete:CASCADE;" json:"budget_output"`
	SAPCI_Tag      *string            `json:"sap_ci_tag,omitempty"` // Sesuai framework: tagging CI di SAP
}

// TableName returns the table name for BudgetComponentModel
func (b *BudgetComponentModel) TableName() string {
	return "budget_components"
}

// BeforeCreate sets the default ID for BudgetComponentModel
func (b *BudgetComponentModel) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		b.ID = uuid.New().String()
	}
	return nil
}
