package models

import (
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ContentCategoryModel struct {
	shared.BaseModel
	Name string `gorm:"type:varchar(255);not null" json:"name"`
	Slug string `gorm:"type:varchar(255);not null;uniqueIndex" json:"slug"`
	Type string `gorm:"type:varchar(255);not null" json:"type"`
}

func (ContentCategoryModel) TableName() string {
	return "content_categories"
}

func (m *ContentCategoryModel) BeforeCreate(tx *gorm.DB) (err error) {
	m.Slug = utils.URLify(m.Name)
	m.ID = uuid.New().String()
	return
}
