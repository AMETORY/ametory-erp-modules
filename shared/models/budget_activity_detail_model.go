package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BudgetActivityDetailModel is a struct for budget activity detail model
type BudgetActivityDetailModel struct {
	shared.BaseModel
	BudgetActivityID *string              `gorm:"type:char(36);index" json:"budget_activity_id"`
	BudgetActivity   *BudgetActivityModel `gorm:"foreignKey:BudgetActivityID;constraint:OnDelete:CASCADE;" json:"budget_activity"`
	Name             string               `gorm:"type:varchar(255);not null" json:"name"`
	Description      string               `gorm:"type:text" json:"description"`
	PlannedAmount    float64              `json:"planned_amount"` // Anggaran yang direncanakan untuk detail ini
	CoAID            *string              `json:"coa_id"`         // FK ke CoA yang relevan
	CoA              *AccountModel        `gorm:"foreignKey:CoAID;constraint:OnDelete:SET NULL;" json:"coa"`
	TransactionID    *string              `json:"transaction_id"` // FK ke transaksi yang relevan
	Transaction      *TransactionModel    `gorm:"foreignKey:TransactionID;constraint:OnDelete:SET NULL;" json:"transaction"`
	ActualAmount     float64              `json:"actual_amount"`         // Jumlah realisasi aktual
	RealizedAt       *time.Time           `json:"realized_at,omitempty"` // Kapan realisasi dicatat (bisa null jika belum ada realisasi)
	RealizedBy       string               `json:"realized_by,omitempty"` // User ID yang mencatat realisasi
}

// TableName returns the table name for BudgetActivityDetailModel
func (b *BudgetActivityDetailModel) TableName() string {
	return "budget_activity_details"
}

// BeforeCreate sets the default ID for BudgetActivityDetailModel
func (b *BudgetActivityDetailModel) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		b.ID = uuid.New().String()
	}
	return nil
}
