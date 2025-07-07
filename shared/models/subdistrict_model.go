package models

import (
	"encoding/json"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SubDistrict struct {
	shared.BaseModel
	Name         string           `gorm:"type:varchar(255)" json:"name"`
	DistrictID   string           `gorm:"size:36" json:"district_id"`
	District     District         `gorm:"foreignKey:DistrictID;references:id" json:"district"`
	Code         string           `gorm:"type:varchar(255);unique" json:"code"`
	Address      string           `json:"address"`
	HeaderLetter string           `json:"header_letter"`
	Logo         string           `json:"logo"`
	Footer       string           `json:"footer"`
	Data         *json.RawMessage `json:"data"`
}

type District struct {
	shared.BaseModel
	Name    string `gorm:"type:varchar(255)" json:"name"`
	Address string `json:"address"`
	CityID  string `gorm:"size:36" json:"city_id"`
	City    City   `gorm:"foreignKey:CityID;references:id" json:"city"`
	Code    string `gorm:"type:varchar(255);unique" json:"code"`
}

type City struct {
	shared.BaseModel
	Name       string   `gorm:"type:varchar(255)" json:"name"`
	ProvinceID string   `gorm:"size:36" json:"province_id"`
	Province   Province `gorm:"foreignKey:ProvinceID;references:id" json:"province"`
	Code       string   `gorm:"type:varchar(255);unique" json:"code"`
}

type Province struct {
	shared.BaseModel
	Name    string `gorm:"type:varchar(255)" json:"name"`
	Address string `json:"address"`
	Code    string `gorm:"type:varchar(255);unique" json:"code"`
}

func (s *SubDistrict) BeforeCreate(tx *gorm.DB) (err error) {
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
