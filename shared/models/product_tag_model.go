package models

import (
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TagModel struct {
	shared.BaseModel
	Name        string        `gorm:"not null" json:"name"`
	Description string        `json:"description,omitempty"`
	Color       string        `gorm:"type:varchar(255);default:'#FFFFFF'" json:"color,omitempty"`
	IconUrl     string        `gorm:"type:varchar(255)" json:"icon_url,omitempty"`
	CompanyID   *string       `json:"company_id,omitempty"`
	Company     *CompanyModel `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE" json:"company,omitempty"`
}

func (TagModel) TableName() string {
	return "tags"
}

func (p *TagModel) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}

type ProductTag struct {
	ProductModelID string `gorm:"primaryKey;uniqueIndex:product_tags_product_id_tag_id_key"`
	TagModelID     string `gorm:"primaryKey;uniqueIndex:product_tags_product_id_tag_id_key"`
}
type VariantTag struct {
	VariantModelID string `gorm:"primaryKey;uniqueIndex:variant_tags_variant_id_tag_id_key"`
	TagModelID     string `gorm:"primaryKey;uniqueIndex:variant_tags_variant_id_tag_id_key"`
}
