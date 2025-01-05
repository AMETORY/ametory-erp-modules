package product

import (
	"github.com/AMETORY/ametory-erp-modules/company"
	"github.com/AMETORY/ametory-erp-modules/inventory/brand"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MasterProductModel struct {
	utils.BaseModel
	Name        string `gorm:"not null"`
	Description string
	SKU         string               `gorm:"type:varchar(255)"`
	Barcode     string               `gorm:"type:varchar(255)"`
	Price       float64              `gorm:"not null;default:0"`
	CompanyID   string               `json:"company_id"`
	Company     company.CompanyModel `gorm:"foreignKey:CompanyID"`
	CategoryID  string
	Category    ProductCategoryModel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;foreignKey:CategoryID"`
	BrandID     string
	Brand       brand.BrandModel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;foreignKey:BrandID"`
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
