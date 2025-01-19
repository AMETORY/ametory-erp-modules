package models

import (
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProductModel struct {
	shared.BaseModel
	Name            string                `gorm:"not null" json:"name,omitempty"`
	Description     *string               `json:"description,omitempty"`
	SKU             *string               `gorm:"type:varchar(255)" json:"sku,omitempty"`
	Barcode         *string               `gorm:"type:varchar(255)" json:"barcode,omitempty"`
	Price           float64               `gorm:"not null;default:0" json:"price,omitempty"`
	CompanyID       *string               `json:"company_id,omitempty"`
	Company         *CompanyModel         `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE" json:"company,omitempty"`
	DistributorID   *string               `gorm:"foreignKey:DistributorID;references:ID;constraint:OnDelete:CASCADE" json:"distributor_id,omitempty"`
	Distributor     *DistributorModel     `gorm:"foreignKey:DistributorID;constraint:OnDelete:CASCADE" json:"distributor,omitempty"`
	MasterProductID *string               `json:"master_product_id,omitempty"`
	MasterProduct   *MasterProductModel   `gorm:"foreignKey:MasterProductID;constraint:OnDelete:CASCADE" json:"master_product,omitempty"`
	CategoryID      *string               `json:"category_id,omitempty"`
	Category        *ProductCategoryModel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:CategoryID" json:"category,omitempty"`
	Prices          []PriceModel          `gorm:"-" json:"prices,omitempty"`
	Brand           *BrandModel           `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:BrandID" json:"brand,omitempty"`
	BrandID         *string               `json:"brand_id,omitempty"`
	ProductImages   []FileModel           `gorm:"-" json:"product_images,omitempty"`
	TotalStock      float64               `gorm:"-" json:"total_stock,omitempty"`
	Status          string                `gorm:"type:VARCHAR(20);default:'ACTIVE'" json:"status,omitempty"`
	Merchants       []*MerchantModel      `gorm:"many2many:product_merchants;constraint:OnDelete:CASCADE;" json:"merchants,omitempty"`
}

func (ProductModel) TableName() string {
	return "products"
}

func (p *ProductModel) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}
