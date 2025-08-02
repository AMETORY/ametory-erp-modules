package models

import (
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MedicalStaffModel struct {
	shared.BaseModel
	Name string `json:"name,omitempty"`
	Type string `json:"type,omitempty"` // "doctor" or "nurse"
}

func (m MedicalStaffModel) TableName() string {
	return "medical_staffs"
}

func (m *MedicalStaffModel) BeforeCreate(tx *gorm.DB) error {
	m.ID = uuid.NewString()
	return nil
}

type NurseModel struct {
	shared.BaseModel
	MedicalStaffID    string            `json:"medical_staff_id,omitempty"`
	MedicalStaffModel MedicalStaffModel `gorm:"foreignKey:MedicalStaffID;constraint:OnDelete:CASCADE;"`
	Certification     string            `json:"certification,omitempty"`
}

func (n NurseModel) TableName() string {
	return "nurses"
}

func (n *NurseModel) BeforeCreate(tx *gorm.DB) error {

	if n.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return nil
}
