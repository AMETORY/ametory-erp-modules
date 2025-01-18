package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PriceModel struct {
	shared.BaseModel
	Amount          float64            `gorm:"not null" json:"amount"`
	Currency        string             `gorm:"type:varchar(3);not null" json:"currency"` // ISO 4217 currency code
	ProductID       string             `json:"product_id"`
	Product         ProductModel       `gorm:"foreignKey:ProductID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
	PriceCategoryID string             `json:"price_category_id"`
	PriceCategory   PriceCategoryModel `gorm:"foreignKey:PriceCategoryID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"price_category"`
	EffectiveDate   time.Time          `json:"effective_date"`
	MinQuantity     float64            `gorm:"not null;default:0" json:"min_quantity"`
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
	shared.BaseModel
	Amount          float64            `gorm:"not null" json:"amount"`
	Currency        string             `gorm:"type:varchar(3);not null" json:"currency"` // ISO 4217 currency code
	MasterProductID string             `json:"master_product_id"`
	MasterProduct   MasterProductModel `gorm:"foreignKey:MasterProductID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
	PriceCategoryID string             `json:"price_category_id"`
	PriceCategory   PriceCategoryModel `gorm:"foreignKey:PriceCategoryID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"price_category"`
	EffectiveDate   time.Time          `json:"effective_date"`
	MinQuantity     float64            `gorm:"not null;default:0" json:"min_quantity"`
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
