package models

import "github.com/AMETORY/ametory-erp-modules/shared"

type TreatmentQueueModel struct {
	shared.BaseModel
	PatientID      *string            `gorm:"type:char(36);index" json:"patient_id"`
	Patient        *PatientModel      `gorm:"foreignKey:PatientID;references:ID" json:"patient"`
	SubFacilityID  *string            `gorm:"type:char(36);index" json:"sub_facility_id"`
	SubFacility    *SubFacilityModel  `gorm:"foreignKey:SubFacilityID;references:ID" json:"sub_facility"`
	MedicalStaffID *string            `gorm:"type:char(36);index" json:"medical_staff_id"`
	MedicalStaff   *MedicalStaffModel `gorm:"foreignKey:MedicalStaffID;references:ID" json:"medical_staff"`
	QueueNumber    string             `gorm:"type:string;not null" json:"queue_number"`
	QueueStatus    string             `gorm:"type:varchar(255);not null" json:"queue_status"`
}
