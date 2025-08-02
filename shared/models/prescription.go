package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Prescription stores the drug information given by the doctor.
type Prescription struct {
	shared.BaseModel
	ConsultationID    string             `json:"consultation_id"`
	Consultation      Consultation       `gorm:"foreignKey:ConsultationID;references:ID" json:"consultation"`
	PatientID         string             `json:"patient_id" gorm:"type:char(36);index"`
	Patient           PatientModel       `gorm:"foreignKey:PatientID;references:ID" json:"patient"`
	DoctorID          string             `json:"doctor_id" gorm:"type:char(36);index"`
	Doctor            Doctor             `gorm:"foreignKey:DoctorID;references:ID" json:"doctor"`
	Date              time.Time          `json:"date"`
	MedicationDetails []MedicationDetail `json:"medication_details" gorm:"foreignKey:PrescriptionID;references:ID"`
}

func (m *Prescription) BeforeCreate(tx *gorm.DB) (err error) {

	if m.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}

// MedicationDetail represents each drug item in a prescription.
type MedicationDetail struct {
	shared.BaseModel
	PrescriptionID string  `json:"prescription_id" gorm:"type:char(36);index"`
	MedicationName string  `json:"medication_name"`
	Dosage         string  `json:"dosage"`
	Quantity       float64 `json:"quantity"`
	Unit           string  `json:"unit"` // Added to specify the unit of the medication (e.g., "tablet", "cc").
	Instructions   string  `json:"instructions"`
}

func (m *MedicationDetail) BeforeCreate(tx *gorm.DB) (err error) {

	if m.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}
