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
	BudgetActivityID *string                  `gorm:"type:char(36);index" json:"budget_activity_id"`
	BudgetActivity   *BudgetActivityModel     `gorm:"foreignKey:BudgetActivityID;constraint:OnDelete:CASCADE;" json:"budget_activity"`
	Name             string                   `gorm:"type:varchar(255);not null" json:"name"`
	Description      string                   `gorm:"type:text" json:"description"`
	PlannedAmount    float64                  `json:"planned_amount"` // Anggaran yang direncanakan untuk detail ini
	CoAID            *string                  `json:"coa_id"`         // FK ke CoA yang relevan
	CoA              *AccountModel            `gorm:"foreignKey:CoAID;constraint:OnDelete:SET NULL;" json:"coa"`
	ActualAmount     float64                  `json:"actual_amount"` // Jumlah realisasi aktual
	Realizations     []BudgetRealizationModel `gorm:"foreignKey:BudgetActivityDetailID;constraint:OnDelete:CASCADE;" json:"realizations"`
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

type BudgetRealizationModel struct {
	shared.BaseModel
	BudgetActivityDetailID *string                    `gorm:"type:char(36);index" json:"budget_activity_detail_id"`
	BudgetActivityDetail   *BudgetActivityDetailModel `gorm:"foreignKey:BudgetActivityDetailID;constraint:OnDelete:CASCADE;" json:"budget_activity_detail"`
	TransactionID          *string                    `gorm:"type:char(36);index" json:"transaction_id"`
	Transaction            *TransactionModel          `gorm:"foreignKey:TransactionID;constraint:OnDelete:CASCADE;" json:"transaction"` // Cascade jika transaksi dihapus, realisasi juga hilang
	Amount                 float64                    `json:"amount"`                                                                   // Jumlah yang direalisasikan DARI transaksi ini UNTUK detail ini
	Notes                  string                     `gorm:"type:text" json:"notes,omitempty"`                                         // Catatan spesifik untuk realisasi ini
	RecordedBy             string                     `json:"recorded_by"`                                                              // User ID yang mencatat realisasi ini
	RealizedAt             time.Time                  `json:"realized_at"`                                                              // Kapan realisasi ini dicatat
}

// TableName returns the table name for BudgetRealizationModel
func (b *BudgetRealizationModel) TableName() string {
	return "budget_realizations"
}

// BeforeCreate sets the default ID for BudgetRealizationModel
func (b *BudgetRealizationModel) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		b.ID = uuid.New().String()
	}
	return nil
}
