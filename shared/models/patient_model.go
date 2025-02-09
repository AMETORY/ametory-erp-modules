package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PatientModel struct {
	shared.BaseModel
	FullName             string     `json:"full_name,omitempty" gorm:"type:varchar(255);index"`
	IdentityCardNumber   string     `json:"identity_card_number,omitempty" gorm:"type:varchar(20);index"`
	SocialSecurityNumber string     `json:"social_security_number,omitempty" gorm:"type:varchar(20);index"`
	Title                *string    `json:"title,omitempty" gorm:"type:varchar(255)"`
	Address              string     `json:"address,omitempty"`
	DateOfBirth          *time.Time `json:"date_of_birth,omitempty"`
	PhoneNumber          string     `json:"phone_number,omitempty" gorm:"type:varchar(20)"`
	Email                string     `json:"email,omitempty" gorm:"type:varchar(255)"`
	ProvinceID           *string    `json:"province_id,omitempty" gorm:"type:char(2);index;constraint:OnDelete:SET NULL;"`
	RegencyID            *string    `json:"regency_id,omitempty" gorm:"type:char(4);index;constraint:OnDelete:SET NULL;"`
	DistrictID           *string    `json:"district_id,omitempty" gorm:"type:char(6);index;constraint:OnDelete:SET NULL;"`
	VillageID            *string    `json:"village_id,omitempty" gorm:"type:char(10);index;constraint:OnDelete:SET NULL;"`
}

func (PatientModel) TableName() string {
	return "patients"
}

func (p *PatientModel) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return nil
}
