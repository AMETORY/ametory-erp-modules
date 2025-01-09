package stockmovement

import (
	"github.com/AMETORY/ametory-erp-modules/inventory/product"
	"github.com/AMETORY/ametory-erp-modules/inventory/warehouse"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MovementType string

const (
	MovementTypeIn       MovementType = "IN"       // Stok masuk
	MovementTypeOut      MovementType = "OUT"      // Stok keluar
	MovementTypeTransfer MovementType = "TRANSFER" // Transfer stok
	MovementTypeAdjust   MovementType = "ADJUST"   // Penyesuaian stok
)

type StockMovementModel struct {
	utils.BaseModel
	Description   string                   `gorm:"null" json:"description"`
	ProductID     string                   `gorm:"not null" json:"product_id"` // Relasi ke product
	Product       product.ProductModel     `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;foreignKey:ProductID" json:"product"`
	WarehouseID   string                   `gorm:"not null" json:"warehouse_id"` // Relasi ke warehouse
	Warehouse     warehouse.WarehouseModel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;foreignKey:WarehouseID" json:"warehouse"`
	MerchantID    *string                  `gorm:"null" json:"merchant_id"`    // Relasi ke warehouse
	DistributorID *string                  `gorm:"null" json:"distributor_id"` // Relasi ke warehouse
	CompanyID     *string                  `gorm:"null" json:"company_id"`     // Relasi ke warehouse
	Quantity      float64                  `gorm:"not null" json:"quantity"`   // Jumlah stok (positif untuk IN, negatif untuk OUT)
	Type          MovementType             `gorm:"not null" json:"type"`       // Jenis pergerakan (IN, OUT, TRANSFER, ADJUST)
	ReferenceID   string                   `json:"reference_id"`               // ID referensi (misalnya, ID pembelian, penjualan, dll.)
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

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&StockMovementModel{})
}
