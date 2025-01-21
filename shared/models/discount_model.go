package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DiscountType string

const (
	DiscountPercentage DiscountType = "PERCENTAGE" // Diskon persentase
	DiscountAmount     DiscountType = "AMOUNT"     // Diskon nominal
)

type DiscountModel struct {
	shared.BaseModel
	ProductID string        `json:"product_id,omitempty"` // ID produk yang didiskon
	Product   *ProductModel `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE" json:"product,omitempty"`
	Type      DiscountType  `gorm:"not null" json:"type,omitempty"`       // Jenis diskon (PERCENTAGE atau AMOUNT)
	Value     float64       `gorm:"not null" json:"value,omitempty"`      // Nilai diskon (persentase atau nominal)
	StartDate time.Time     `gorm:"not null" json:"start_date,omitempty"` // Tanggal mulai diskon
	EndDate   *time.Time    `json:"end_date,omitempty"`
	IsActive  bool          `gorm:"not null" json:"is_active,omitempty"` // Status diskon (aktif atau tidak)
	Notes     string        `json:"notes"`                               // Catatan diskon
}

func (DiscountModel) TableName() string {
	return "discounts"
}

func (d *DiscountModel) BeforeCreate(tx *gorm.DB) (err error) {
	if d.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}
