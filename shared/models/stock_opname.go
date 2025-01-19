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
	WarehouseID string              `gorm:"not null"` // ID gudang
	Status      StockOpnameStatus   `gorm:"not null"` // Status stock opname
	OpnameDate  time.Time           `gorm:"not null"` // Tanggal stock opname
	Notes       string              // Catatan tambahan
	Details     []StockOpnameDetail `gorm:"foreignKey:StockOpnameID;constraint:OnDelete:CASCADE"` // Detail produk
}

type StockOpnameDetail struct {
	shared.BaseModel
	StockOpnameID string  `gorm:"not null"` // ID stock opname header
	ProductID     string  `gorm:"not null"` // ID produk
	Quantity      float64 `gorm:"not null"` // Jumlah stok fisik
	SystemQty     float64 `gorm:"not null"` // Jumlah stok di sistem
	Difference    float64 `gorm:"not null"` // Selisih stok (Quantity - SystemQty)
	Notes         string  // Catatan tambahan
}
