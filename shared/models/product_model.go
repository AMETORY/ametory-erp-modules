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
	Company         *CompanyModel         `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	DistributorID   *string               `gorm:"foreignKey:DistributorID;references:ID" json:"distributor_id,omitempty"`
	Distributor     *DistributorModel     `gorm:"foreignKey:DistributorID" json:"distributor,omitempty"`
	MasterProductID *string               `json:"master_product_id,omitempty"`
	MasterProduct   *MasterProductModel   `gorm:"foreignKey:MasterProductID" json:"master_product,omitempty"`
	CategoryID      *string               `json:"category_id,omitempty"`
	Category        *ProductCategoryModel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;foreignKey:CategoryID" json:"category,omitempty"`
	Prices          []PriceModel          `gorm:"-" json:"prices,omitempty"`
	Brand           *BrandModel           `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;foreignKey:BrandID" json:"brand,omitempty"`
	BrandID         *string               `json:"brand_id,omitempty"`
	ProductImages   []shared.FileModel    `gorm:"-" json:"product_images,omitempty"`
	TotalStock      float64               `gorm:"-" json:"total_stock,omitempty"`
	Status          string                `gorm:"type:VARCHAR(20);default:'ACTIVE'" json:"status,omitempty"`
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
