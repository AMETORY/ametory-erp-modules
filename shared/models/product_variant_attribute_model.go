package models

import (
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type VariantProductAttributeModel struct {
	shared.BaseModel
	VariantID   string                `gorm:"index" json:"variant_id,omitempty"`
	AttributeID string                `gorm:"index" json:"attribute_id,omitempty"`
	Attribute   ProductAttributeModel `gorm:"foreignKey:AttributeID;constraint:OnDelete:CASCADE" json:"attribute,omitempty"`
	Value       string                `gorm:"type:varchar(255)" json:"value,omitempty"`
}

func (VariantProductAttributeModel) TableName() string {
	return "product_variant_attributes"
}

func (v *VariantProductAttributeModel) BeforeCreate(tx *gorm.DB) error {
	if v.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return nil
}
