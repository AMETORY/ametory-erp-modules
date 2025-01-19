package models

import (
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProductAttributeModel struct {
	shared.BaseModel
	Name string
}

func (ProductAttributeModel) TableName() string {
	return "product_attributes"
}

func (p *ProductAttributeModel) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}
