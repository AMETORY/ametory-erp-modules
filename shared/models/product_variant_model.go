package models

import (
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type VariantModel struct {
	shared.BaseModel
	ProductID  string                         `gorm:"index" json:"product_id,omitempty"`
	Product    ProductModel                   `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	SKU        string                         `gorm:"type:varchar(255);not null" json:"sku,omitempty"`
	Barcode    *string                        `gorm:"type:varchar(255)" json:"barcode,omitempty"`
	Price      float64                        `gorm:"not null;default:0" json:"price,omitempty"`
	Attributes []VariantProductAttributeModel `gorm:"foreignKey:VariantID;constraint:OnDelete:CASCADE" json:"attributes,omitempty"`
}

func (VariantModel) TableName() string {
	return "product_variants"
}

func (v *VariantModel) BeforeCreate(tx *gorm.DB) error {
	if v.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return nil
}
