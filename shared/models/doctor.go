package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Doctor struct {
	shared.BaseModel
	Name             string                `json:"name"`
	Title            string                `json:"title"`
	SuffixTitle      string                `json:"suffix_title"`
	PhoneNumber      string                `json:"phone_number"`
	STRNumber        string                `json:"str_number"`
	SIPNumber        string                `json:"sip_number"`
	AvailableSlots   []DoctorSchedule      `gorm:"foreignKey:DoctorID;references:ID" json:"available_slots"`
	Reviews          string                `json:"reviews"`
	FullName         string                `json:"full_name" gorm:"-"`
	Avatar           *FileModel            `json:"avatar" gorm:"-"`
	SpecializationID *string               `json:"specialization_id,omitempty" gorm:"type:char(36);index"`
	Specialization   *DoctorSpecialization `gorm:"foreignKey:SpecializationID;references:ID" json:"specialization,omitempty"`
	Email            string                `json:"email"`
	Address          string                `json:"address"`
}

func (m *Doctor) BeforeCreate(tx *gorm.DB) (err error) {

	if m.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}

func (u *Doctor) AfterFind(tx *gorm.DB) error {
	file := FileModel{}
	err := tx.Where("ref_id = ? and ref_type = ?", u.ID, "doctor-avatar").Order("created_at desc").First(&file).Error
	if err == nil {
		u.Avatar = &file
	}
	name := u.Name
	if u.Title != "" {
		name = u.Title + " " + u.Name
	}
	if u.SuffixTitle != "" {
		name = name + ", " + u.SuffixTitle
	}
	u.FullName = name
	return nil
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

type DoctorSpecialization struct {
	shared.BaseModel
	Code        string `json:"code" gorm:"uniqueIndex"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (m *DoctorSpecialization) BeforeCreate(tx *gorm.DB) (err error) {

	if m.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}
