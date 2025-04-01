package models

import (
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProductCategoryModel struct {
	shared.BaseModel
	Name        string         `gorm:"unique;not null" json:"name"`
	Description string         `json:"description"`
	Color       string         `gorm:"type:varchar(255);default:'#94CFCD'" json:"color"`
	IconUrl     string         `gorm:"type:varchar(255)" json:"icon_url"`
	Products    []ProductModel `gorm:"foreignKey:CategoryID" json:"products"`
	CompanyID   *string        `json:"company_id,omitempty"`
	Company     *CompanyModel  `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE" json:"company,omitempty"`
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
