package models

import (
	"sort"
	"strings"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type VariantModel struct {
	shared.BaseModel
	ProductID   string                         `gorm:"index" json:"product_id,omitempty"`
	Product     ProductModel                   `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	SKU         string                         `gorm:"type:varchar(255);not null" json:"sku,omitempty"`
	Barcode     *string                        `gorm:"type:varchar(255)" json:"barcode,omitempty"`
	Price       float64                        `gorm:"not null;default:0" json:"price,omitempty"`
	Attributes  []VariantProductAttributeModel `gorm:"foreignKey:VariantID;constraint:OnDelete:CASCADE" json:"attributes,omitempty"`
	DisplayName string                         `gorm:"type:varchar(255)" json:"display_name,omitempty"`
}

func (VariantModel) TableName() string {
	return "product_variants"
}

func (v *VariantModel) BeforeCreate(tx *gorm.DB) error {
	if v.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	v.GenerateDisplayName(tx)
	return nil
}

func (v *VariantModel) GenerateDisplayName(tx *gorm.DB) {
	var productName string
	attributeValues := make([]string, len(v.Attributes))

	var product ProductModel
	if err := tx.Preload("Brand", func(db *gorm.DB) *gorm.DB {
		return db.Select("name", "id")
	}).Where("id = ?", v.ProductID).Select("id", "name", "brand_id").First(&product).Error; err != nil {
		return
	}
	productName = product.Name

	// Sort attributes by priority
	sort.SliceStable(v.Attributes, func(i, j int) bool {
		return v.Attributes[i].Attribute.Priority < v.Attributes[j].Attribute.Priority
	})

	// Collect attribute values
	for i, attr := range v.Attributes {
		attributeValues[i] = attr.Value
	}

	var displayNames []string
	if product.Brand != nil {
		displayNames = append(displayNames, product.Brand.Name)
	}
	displayNames = append(displayNames, productName)
	displayNames = append(displayNames, strings.Join(attributeValues, " "))

	// Generate display name
	v.DisplayName = strings.Join(displayNames, " ")
}
