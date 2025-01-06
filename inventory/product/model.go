package product

import (
	"github.com/AMETORY/ametory-erp-modules/company"
	"github.com/AMETORY/ametory-erp-modules/inventory/brand"
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProductModel struct {
	utils.BaseModel
	Name            string               `gorm:"not null" json:"name"`
	Description     string               `json:"description"`
	SKU             string               `gorm:"type:varchar(255)" json:"sku"`
	Barcode         string               `gorm:"type:varchar(255)" json:"barcode"`
	Price           float64              `gorm:"not null;default:0" json:"price"`
	CompanyID       string               `json:"company_id"`
	Company         company.CompanyModel `gorm:"foreignKey:CompanyID"`
	DistributorID   *string              `json:"distributor_id"`
	Distributor     interface{}          `gorm:"foreignKey:DistributorID"`
	MasterProductID string               `json:"master_product_id"`
	MasterProduct   MasterProductModel   `gorm:"foreignKey:MasterProductID"`
	CategoryID      string               `json:"category_id"`
	Category        ProductCategoryModel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;foreignKey:CategoryID"`
	Prices          []PriceModel         `gorm:"-" json:"prices"`
	Brand           brand.BrandModel     `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;foreignKey:BrandID"`
	BrandID         string               `json:"brand_id"`
	ProductImages   []shared.FileModel   `json:"product_images"`
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

type ProductCategoryModel struct {
	utils.BaseModel
	Name        string `gorm:"unique;not null"`
	Description string
}

func (ProductCategoryModel) TableName() string {
	return "product_categories"
}

func (p *ProductCategoryModel) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&ProductModel{}, &ProductCategoryModel{}, &MasterProductModel{}, &PriceCategoryModel{}, &PriceModel{}, &MasterProductPriceModel{})
}
