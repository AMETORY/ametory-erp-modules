package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MedicalAppointmentModel adalah model database untuk appointment
type MedicalAppointmentModel struct {
	shared.BaseModel
	PatientID      string            `gorm:"type:char(36);index" json:"patient_id"`
	Patient        PatientModel      `gorm:"foreignKey:PatientID;references:ID" json:"patient"`
	MedicalStaffID string            `gorm:"type:char(36);index" json:"medical_staff_id"`
	MedicalStaff   MedicalStaffModel `gorm:"foreignKey:MedicalStaffID;references:ID" json:"medical_staff"`
	Schedule       time.Time         `gorm:"autoCreateTime" json:"schedule"`
	Status         string            `gorm:"default:pending" json:"status"`
	SubFacilityID  string            `gorm:"type:char(36);index" json:"sub_facility_id"`
	SubFacility    SubFacilityModel  `gorm:"foreignKey:SubFacilityID;references:ID" json:"sub_facility"`
}

func (MedicalAppointmentModel) TableName() string {
	return "medical_appointments"
}

func (m *MedicalAppointmentModel) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return nil
}
