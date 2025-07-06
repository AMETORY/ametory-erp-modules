package models

import (
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Citizen struct {
	shared.BaseModel
	FullName string `json:"full_name,omitempty"`
	NIK      string `gorm:"uniqueIndex;index" json:"nik,omitempty"`
	Address  string `json:"address,omitempty"`
	Phone    string `json:"phone,omitempty"`
}

func (c *Citizen) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}
