package models

import (
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PermitTemplate struct {
	shared.BaseModel
	Name        string `json:"name"`
	Description string `json:"description"`
	Slug        string `json:"slug"`
}

func (p *PermitTemplate) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return
}
