package models

import (
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SubFacilityModel struct {
	shared.BaseModel
	FacilityID   string              `gorm:"type:char(36);index" json:"facility_id"`
	Name         string              `gorm:"type:varchar(255);not null" json:"name"`
	PhoneNumber  string              `gorm:"type:varchar(255)" json:"phone_number,omitempty"`
	Email        string              `gorm:"type:varchar(255)" json:"email,omitempty"`   // TODO: validate email
	Website      string              `gorm:"type:varchar(255)" json:"website,omitempty"` // TODO: validate website
	MedicalStaff []MedicalStaffModel `gorm:"many2many:sub_facility_medical_staffs;constraint:OnDelete:CASCADE;" json:"medical_staff,omitempty"`
}

func (s SubFacilityModel) TableName() string {
	return "sub_facilities"
}

func (s SubFacilityModel) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return nil
}

type SubFacilityStaff struct {
	shared.BaseModel
	SubFacilityModelID  string `gorm:"type:char(36);index" json:"sub_facility_id"`
	MedicalStaffModelID string `gorm:"type:char(36);index" json:"medical_staff_id"`
}
