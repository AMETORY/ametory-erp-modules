package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MovementType string

const (
	MovementTypeIn       MovementType = "IN"       // Stok masuk
	MovementTypeOut      MovementType = "OUT"      // Stok keluar
	MovementTypeSale     MovementType = "SALE"     // Stok keluar
	MovementTypePurchase MovementType = "PURCHASE" // Stok keluar
	MovementTypeReturn   MovementType = "RETURN"   // Stok keluar
	MovementTypeTransfer MovementType = "TRANSFER" // Transfer stok
	MovementTypeAdjust   MovementType = "ADJUST"   // Penyesuaian stok
)

type StockMovementModel struct {
	shared.BaseModel
	Date              time.Time           `gorm:"not null" json:"date"` // Tanggal pergerakan stok
	Description       string              `gorm:"null" json:"description"`
	ProductID         string              `gorm:"not null" json:"product_id"` // Relasi ke product
	Product           ProductModel        `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:ProductID" json:"product"`
	VariantID         *string             `json:"variant_id,omitempty"`
	Variant           *VariantModel       `gorm:"foreignKey:VariantID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
	SourceWarehouseID string              `gorm:"-" json:"source_warehouse_id"` // Relasi ke warehouse
	WarehouseID       string              `gorm:"not null" json:"warehouse_id"` // Relasi ke warehouse
	Warehouse         WarehouseModel      `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:WarehouseID" json:"warehouse"`
	MerchantID        *string             `gorm:"null" json:"merchant_id"` // Relasi ke merchant
	Merchant          *MerchantModel      `gorm:"foreignKey:MerchantID;constraint:OnDelete:CASCADE" json:"merchant"`
	DistributorID     *string             `gorm:"null" json:"distributor_id"` // Relasi ke distributor
	Distributor       *DistributorModel   `gorm:"foreignKey:DistributorID;constraint:OnDelete:CASCADE" json:"distributor"`
	CompanyID         *string             `gorm:"null" json:"company_id"` // Relasi ke company
	Company           *CompanyModel       `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE" json:"company"`
	Quantity          float64             `gorm:"not null" json:"quantity"` // Jumlah stok (positif untuk IN, negatif untuk OUT)
	Value             float64             `gorm:"not null;default:1" json:"value"`
	Type              MovementType        `gorm:"not null" json:"type"` // Jenis pergerakan (IN, OUT, TRANSFER, ADJUST)
	ReferenceID       string              `json:"reference_id"`         // ID referensi (misalnya, ID pembelian, penjualan, dll.)
	ReferenceType     *string             `json:"reference_type"`       // Jenis referensi (misalnya, PURCHASE, SALE, TRANSFER, ADJUST)
	SecondaryRefID    *string             `json:"secondary_ref_id,omitempty"`
	SecondaryRefType  *string             `gorm:"secondary_ref_type" json:"secondary_ref_type,omitempty"`
	UnitID            *string             `json:"unit_id,omitempty"` // Relasi ke unit
	Unit              *UnitModel          `gorm:"foreignKey:UnitID;constraint:OnDelete:CASCADE" json:"unit,omitempty"`
	SalesRef          *SalesModel         `gorm:"-" json:"sales_ref,omitempty"`
	PurchaseRef       *PurchaseOrderModel `gorm:"-" json:"purchase_ref,omitempty"`
}

func (StockMovementModel) TableName() string {
	return "stock_movements"
}

func (p *StockMovementModel) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}
