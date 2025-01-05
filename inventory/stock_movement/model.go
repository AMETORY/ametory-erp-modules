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
	ProductID   string                   `gorm:"not null"` // Relasi ke product
	Product     product.ProductModel     `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;foreignKey:ProductID"`
	WarehouseID string                   `gorm:"not null"` // Relasi ke warehouse
	Warehouse   warehouse.WarehouseModel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;foreignKey:WarehouseID"`
	MerchantID  *string                  `gorm:"not null"` // Relasi ke warehouse
	Merchant    interface{}              `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;foreignKey:MerchantID"`
	Quantity    float64                  `gorm:"not null"` // Jumlah stok (positif untuk IN, negatif untuk OUT)
	Type        MovementType             `gorm:"not null"` // Jenis pergerakan (IN, OUT, TRANSFER, ADJUST)
	ReferenceID string                   // ID referensi (misalnya, ID pembelian, penjualan, dll.)
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
