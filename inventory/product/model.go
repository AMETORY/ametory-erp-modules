package product

import (
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProductModel struct {
	utils.BaseModel
	Name        string `gorm:"not null"`
	Description string
	SKU         string  `gorm:"type:varchar(255)"`
	Barcode     string  `gorm:"type:varchar(255)"`
	Price       float64 `gorm:"not null;default:0"`
	CategoryID  string
	Category    ProductCategoryModel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;foreignKey:CategoryID"`
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
	return db.AutoMigrate(&ProductModel{}, &ProductCategoryModel{})
}
