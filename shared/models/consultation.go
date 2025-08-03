package models

import (
	"encoding/json"
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Consultation represents a consultation session.

// Consultation represents a consultation session.
type Consultation struct {
	shared.BaseModel
	PatientID        string            `json:"patient_id" gorm:"type:char(36);index"`
	Patient          PatientModel      `gorm:"foreignKey:PatientID;references:ID" json:"patient"`
	DoctorID         string            `json:"doctor_id" gorm:"type:char(36);index"`
	Doctor           Doctor            `gorm:"foreignKey:DoctorID;references:ID" json:"doctor"`
	ScheduleID       string            `json:"schedule_id"`
	Schedule         DoctorSchedule    `gorm:"foreignKey:ScheduleID;references:ID" json:"schedule"`
	ConsultationType string            `json:"consultation_type"`
	MeetingURL       *string           `json:"meeting_url,omitempty"`
	Status           string            `json:"status"`
	PaymentID        string            `json:"payment_id"`
	ScreeningID      *string           `json:"screening_id" gorm:"type:varchar(36)"`
	Screening        *InitialScreening `gorm:"foreignKey:ScreeningID;references:ID" json:"screening"`
}

// BeforeCreate is the hook for creating a new Consultation.
func (c *Consultation) BeforeCreate(tx *gorm.DB) (err error) {
	if c.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}

type InitialScreening struct {
	ID            string          `json:"id"`
	PatientID     string          `json:"patient_id" gorm:"type:char(36);index"`
	Patient       PatientModel    `gorm:"foreignKey:PatientID;references:ID" json:"patient"`
	MainComplaint string          `json:"main_complaint"`
	Symptoms      json.RawMessage `json:"symptoms"`
	SubmittedAt   time.Time       `json:"submitted_at"`
	Status        string          `json:"status"` // e.g., "Pending", "Completed"
}

// BeforeCreate is the hook for creating a new InitialScreening.
func (is *InitialScreening) BeforeCreate(tx *gorm.DB) (err error) {
	if is.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}

// Payment represents a consultation payment transaction.
type ConsultationPayment struct {
	shared.BaseModel
	ConsultationID    string       `json:"consultation_id"`
	Consultation      Consultation `gorm:"foreignKey:ConsultationID;references:ID" json:"consultation"`
	PatientID         string       `json:"patient_id" gorm:"type:char(36);index"`
	Patient           PatientModel `gorm:"foreignKey:PatientID;references:ID" json:"patient"`
	Amount            float64      `json:"amount"`
	PlatformFee       float64      `json:"platform_fee"`
	PaymentGatewayFee float64      `json:"payment_gateway_fee"`
	DoctorPayout      float64      `json:"doctor_payout"`
	Method            string       `json:"method"`
	PaymentStatus     string       `json:"payment_status"`
	PaymentTime       time.Time    `json:"payment_time"`
}

// BeforeCreate is the hook for creating a new ConsultationPayment.
func (cp *ConsultationPayment) BeforeCreate(tx *gorm.DB) (err error) {
	if cp.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}
