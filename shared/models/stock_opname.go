package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
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
	CompanyID         *string             `json:"company_id,omitempty"`                                                           // ID perusahaan
	Company           *CompanyModel       `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE" json:"company,omitempty"`      // Relasi ke perusahaan
	WarehouseID       string              `gorm:"not null" json:"warehouse_id"`                                                   // ID gudang
	Warehouse         WarehouseModel      `gorm:"foreignKey:WarehouseID;constraint:OnDelete:CASCADE" json:"warehouse,omitempty"`  // Relasi ke gudang
	Status            StockOpnameStatus   `gorm:"not null;default:DRAFT" json:"status"`                                           // Status stock opname
	OpnameDate        time.Time           `gorm:"not null" json:"opname_date"`                                                    // Tanggal stock opname
	Notes             string              `json:"notes,omitempty"`                                                                // Catatan tambahan
	Details           []StockOpnameDetail `gorm:"foreignKey:StockOpnameID;constraint:OnDelete:CASCADE" json:"details,omitempty"`  // Detail produk
	CreatedByID       *string             `json:"created_by_id,omitempty"`                                                        // ID user yang membuat stock opname
	CreatedBy         *UserModel          `gorm:"foreignKey:CreatedByID;constraint:OnDelete:CASCADE" json:"created_by,omitempty"` // Relasi ke user yang membuat stock opname
}

func (s *StockOpnameHeader) BeforeCreate(tx *gorm.DB) (err error) {

	if s.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}

type StockOpnameDetail struct {
	shared.BaseModel
	StockOpnameID string        `gorm:"not null" json:"stock_opname_id,omitempty"`                                 // ID stock opname header
	ProductID     string        `json:"product_id,omitempty"`                                                      // ID produk
	Product       ProductModel  `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE" json:"product,omitempty"` // Relasi ke produk
	VariantID     *string       `json:"variant_id,omitempty"`
	Variant       *VariantModel `gorm:"foreignKey:VariantID;constraint:OnDelete:CASCADE" json:"variant,omitempty"` // Relasi ke varian
	Quantity      float64       `gorm:"not null" json:"quantity"`                                                  // Jumlah stok fisik
	SystemQty     float64       `gorm:"not null" json:"system_qty"`                                                // Jumlah stok di sistem
	Difference    float64       `gorm:"not null" json:"difference"`                                                // Selisih stok (Quantity - SystemQty)
	UnitID        *string       `json:"unit_id,omitempty"`                                                         // Relasi ke unit
	Unit          *UnitModel    `gorm:"foreignKey:UnitID;constraint:OnDelete:CASCADE" json:"unit,omitempty"`
	UnitValue     float64       `gorm:"not null;default:1" json:"unit_value,omitempty"`
	UnitPrice     float64       `gorm:"not null" json:"unit_price,omitempty"`
	Notes         string        `json:"notes,omitempty"` // Catatan tambahan
}

func (d *StockOpnameDetail) BeforeCreate(tx *gorm.DB) (err error) {
	if d.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}
