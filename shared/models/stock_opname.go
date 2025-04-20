package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
)

type StockOpnameStatus string

const (
	StatusDraft      StockOpnameStatus = "DRAFT"
	StatusInProgress StockOpnameStatus = "IN_PROGRESS"
	StatusCompleted  StockOpnameStatus = "COMPLETED"
	StatusCancelled  StockOpnameStatus = "CANCELLED"
)

type StockOpnameHeader struct {
	shared.BaseModel
	StockOpnameNumber string              `json:"stock_opname_number,omitempty"`
	CompanyID         *string             ` json:"company_id"`                                          // ID perusahaan
	Company           *CompanyModel       `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE"`     // Relasi ke perusahaan
	WarehouseID       string              `gorm:"not null" json:"warehouse_id"`                         // ID gudang
	Warehouse         WarehouseModel      `gorm:"foreignKey:WarehouseID;constraint:OnDelete:CASCADE"`   // Relasi ke gudang
	Status            StockOpnameStatus   `gorm:"not null" json:"status"`                               // Status stock opname
	OpnameDate        time.Time           `gorm:"not null" json:"opname_date"`                          // Tanggal stock opname
	Notes             string              `json:"notes,omitempty"`                                      // Catatan tambahan
	Details           []StockOpnameDetail `gorm:"foreignKey:StockOpnameID;constraint:OnDelete:CASCADE"` // Detail produk
	CreatedByID       *string             `json:"created_by_id,omitempty"`                              // ID user yang membuat stock opname
	CreatedBy         *UserModel          `gorm:"foreignKey:CreatedByID;constraint:OnDelete:CASCADE"`   // Relasi ke user yang membuat stock opname
}

type StockOpnameDetail struct {
	shared.BaseModel
	StockOpnameID string     `gorm:"not null"` // ID stock opname header
	ProductID     string     `gorm:"not null"` // ID produk
	VariantID     *string    `json:"variant_id,omitempty"`
	Quantity      float64    `gorm:"not null"`          // Jumlah stok fisik
	SystemQty     float64    `gorm:"not null"`          // Jumlah stok di sistem
	Difference    float64    `gorm:"not null"`          // Selisih stok (Quantity - SystemQty)
	UnitID        *string    `json:"unit_id,omitempty"` // Relasi ke unit
	Unit          *UnitModel `gorm:"foreignKey:UnitID;constraint:OnDelete:CASCADE" json:"unit,omitempty"`
	UnitValue     float64    `gorm:"not null;default:1" json:"unit_value"`
	UnitPrice     float64    `gorm:"not null" json:"unit_price"`
	Notes         string     // Catatan tambahan
}
