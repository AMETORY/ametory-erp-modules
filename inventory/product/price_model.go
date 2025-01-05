package product

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PriceModel struct {
	utils.BaseModel
	Amount          float64            `gorm:"not null"`
	Currency        string             `gorm:"type:varchar(3);not null"` // ISO 4217 currency code
	ProductID       string             `json:"product_id"`
	Product         ProductModel       `gorm:"foreignKey:ProductID"`
	PriceCategoryID string             `json:"price_category_id"`
	PriceCategory   PriceCategoryModel `gorm:"foreignKey:PriceCategoryID"`
	EffectiveDate   time.Time          `json:"effective_date"`
	MinQuantity     float64            `gorm:"not null;default:0"`
}

func (p *PriceModel) TableName() string {
	return "product_prices"
}

func (p *PriceModel) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}

type MasterProductPriceModel struct {
	utils.BaseModel
	Amount          float64            `gorm:"not null"`
	Currency        string             `gorm:"type:varchar(3);not null"` // ISO 4217 currency code
	MasterProductID string             `json:"master_product_id"`
	MasterProduct   MasterProductModel `gorm:"foreignKey:MasterProductID"`
	PriceCategoryID string             `json:"price_category_id"`
	PriceCategory   PriceCategoryModel `gorm:"foreignKey:PriceCategoryID"`
	EffectiveDate   time.Time          `json:"effective_date"`
	MinQuantity     float64            `gorm:"not null;default:0"`
}

func (p *MasterProductPriceModel) TableName() string {
	return "master_product_prices"
}

func (p *MasterProductPriceModel) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}
