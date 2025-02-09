package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MedicalRecordModel struct {
	shared.BaseModel
	PatientID      string            `gorm:"type:char(36);index" json:"patient_id"`
	Patient        PatientModel      `gorm:"foreignKey:PatientID;references:ID" json:"patient"`
	MedicalStaffID string            `gorm:"type:char(36);index" json:"medical_staff_id"`
	MedicalStaff   MedicalStaffModel `gorm:"foreignKey:MedicalStaffID;references:ID" json:"medical_staff"`
	SubFacilityID  string            `gorm:"type:char(36);index" json:"sub_facility_id"`
	SubFacility    SubFacilityModel  `gorm:"foreignKey:SubFacilityID;references:ID" json:"sub_facility"`
	Anamnesis      string            `gorm:"type:text" json:"anamnesis"`       // riwayat kesehatan pasien
	Diagnosis      string            `gorm:"type:text" json:"diagnosis"`       // diagnosa pasien
	ChiefComplaint string            `gorm:"type:text" json:"chief_complaint"` // keluhan utama pasien
	Progress       string            `gorm:"type:text" json:"progress"`        // perkembangan pasien
	Medication     string            `gorm:"type:text" json:"medication"`      // obat-obatan yang dikonsumsi pasien
	OtherNotes     string            `gorm:"type:text" json:"other_notes"`     // catatan lain-lain
	Prescription   string            `gorm:"type:text" json:"prescription"`    // resep obat yang diberikan
	VisitDate      time.Time         `gorm:"autoCreateTime" json:"visit_date"` // tanggal kunjungan pasien
}

func (m *MedicalRecordModel) TableName() string {
	return "medical_records"
}

func (m *MedicalRecordModel) BeforeCreate(tx *gorm.DB) (err error) {
	m.ID = uuid.New().String()
	return
}
