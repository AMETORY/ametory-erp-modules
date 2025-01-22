package models

import (
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TagModel struct {
	shared.BaseModel
	Name        string `gorm:"not null;unique" json:"name"`
	Description string `json:"description,omitempty"`
	Color       string `gorm:"type:varchar(255);default:'#FFFFFF'" json:"color,omitempty"`
	IconUrl     string `gorm:"type:varchar(255)" json:"icon_url,omitempty"`
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
