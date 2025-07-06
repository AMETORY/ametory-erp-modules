package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Citizen struct {
	shared.BaseModel
	FullName      string     `json:"full_name,omitempty"`
	NIK           string     `gorm:"uniqueIndex;index" json:"nik,omitempty"`
	Address       string     `json:"address,omitempty"`
	Phone         string     `json:"phone,omitempty"`
	Gender        string     `json:"gender,omitempty" gorm:"type:varchar(255)"`
	BirthDate     *time.Time `json:"birth_date,omitempty" gorm:"type:date"`
	Occupation    string     `json:"occupation,omitempty" gorm:"type:varchar(255)"`
	BirthPlace    string     `json:"birth_place,omitempty" gorm:"type:varchar(255)"`
	Religion      string     `json:"religion,omitempty" gorm:"type:varchar(255)"`
	MaritalStatus string     `json:"marital_status,omitempty" gorm:"type:varchar(255)"`
}

func (c *Citizen) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}
