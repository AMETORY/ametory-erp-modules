package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Doctor struct {
	shared.BaseModel
	Name           string           `json:"name"`
	Title          string           `json:"title"`
	Specialization string           `json:"specialization"`
	STRNumber      string           `json:"str_number"`
	SIPNumber      string           `json:"sip_number"`
	AvailableSlots []DoctorSchedule `gorm:"foreignKey:DoctorID;references:ID" json:"available_slots"`
	Reviews        string           `json:"reviews"`
}

func (m *Doctor) BeforeCreate(tx *gorm.DB) (err error) {

	if m.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}

type DoctorSchedule struct {
	shared.BaseModel
	DoctorID  string    `json:"doctor_id" gorm:"type:char(36);index"`
	Doctor    Doctor    `gorm:"foreignKey:DoctorID;references:ID" json:"doctor"`
	Price     float64   `json:"price"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Status    string    `json:"status"`
}

func (m *DoctorSchedule) BeforeCreate(tx *gorm.DB) (err error) {

	if m.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}
