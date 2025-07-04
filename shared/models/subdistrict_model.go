package models

import (
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Subdistrict struct {
	shared.BaseModel
	Name       string `gorm:"type:varchar(255);not null;unique" json:"name"`
	DistrictID string `gorm:"size:36" json:"district_id"`
	District   District
	Code       string `gorm:"type:varchar(255);unique" json:"code"`
	Address    string `json:"address"`
}

type District struct {
	shared.BaseModel
	Name       string `gorm:"type:varchar(255);not null;unique" json:"name"`
	Address    string `json:"address"`
	ProvinceID string `gorm:"size:36" json:"province_id"`
	Province   Province
	Code       string `gorm:"type:varchar(255);unique" json:"code"`
}

type Province struct {
	shared.BaseModel
	Name    string `gorm:"type:varchar(255);not null;unique" json:"name"`
	Address string `json:"address"`
	Code    string `gorm:"type:varchar(255);unique" json:"code"`
}

func (s *Subdistrict) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	return
}

func (d *District) BeforeCreate(tx *gorm.DB) (err error) {
	if d.ID == "" {
		d.ID = uuid.New().String()
	}
	return
}

func (p *Province) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return
}
