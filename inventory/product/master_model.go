package product

import (
	"github.com/AMETORY/ametory-erp-modules/company"
	"github.com/AMETORY/ametory-erp-modules/inventory/brand"
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MasterProductModel struct {
	utils.BaseModel
	Name          string                    `gorm:"not null" json:"name"`
	Description   string                    `json:"description"`
	SKU           string                    `gorm:"type:varchar(255)" json:"sku"`
	Barcode       string                    `gorm:"type:varchar(255)" json:"barcode"`
	Price         float64                   `gorm:"not null;default:0" json:"price"`
	CompanyID     *string                   `json:"company_id"`
	Company       *company.CompanyModel     `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	CategoryID    *string                   `json:"category_id"`
	Category      *ProductCategoryModel     `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;foreignKey:CategoryID" json:"category,omitempty"`
	Prices        []MasterProductPriceModel `gorm:"-" json:"prices"`
	Brand         *brand.BrandModel         `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;foreignKey:BrandID" json:"brand,omitempty"`
	BrandID       *string                   `json:"brand_id,omitempty"`
	ProductImages []shared.FileModel        `gorm:"-" json:"product_images"`
}

func (MasterProductModel) TableName() string {
	return "master_products"
}

func (m *MasterProductModel) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}
