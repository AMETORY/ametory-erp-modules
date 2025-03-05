package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WorkReport struct {
	shared.BaseModel
	EmployeeID     string          `json:"employee_id"`
	Employee       EmployeeModel   `gorm:"foreignKey:EmployeeID"`
	AttendanceID   *string         `json:"attendance_id"`
	Attendance     AttendanceModel `gorm:"foreignKey:AttendanceID"`
	WorkDate       time.Time       `json:"work_date"`
	WorkTypeID     string          `json:"work_type_id"`
	WorkType       WorkType        `gorm:"foreignKey:WorkTypeID"`                                                            // Foreign key ke tabel WorkType
	UnitsCompleted float64         `json:"units_completed"`                                                                  // Jumlah unit pekerjaan yang diselesaikan
	Status         string          `json:"status" gorm:"type:enum('SUBMITTED', 'APPROVED', 'REJECTED');default:'SUBMITTED'"` // Status laporan: submitted, approved, rejected
	SubmittedAt    time.Time       `json:"submitted_at"`
	ApprovedDate   *time.Time      `json:"approved_date"`
	ApprovedByID   *string         `json:"approved_by_id"`
	ApprovedBy     EmployeeModel   `gorm:"foreignKey:ApprovedByID"`
	CompanyID      string          `json:"company_id"`
	Company        CompanyModel    `gorm:"foreignKey:CompanyID"`
	Pictures       []FileModel     `json:"pictures" gorm:"-"`
	Check          WorkReportCheck `json:"check" gorm:"foreignKey:WorkReportID"`
	Remarks        string          `json:"remarks"`
}

func (WorkReport) TableName() string {
	return "work_reports"
}

func (wr *WorkReport) BeforeCreate(tx *gorm.DB) error {
	if wr.Status == "" {
		wr.Status = "SUBMITTED"
	}
	if wr.SubmittedAt.IsZero() {
		wr.SubmittedAt = time.Now()
	}

	if wr.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return nil
}

type WorkType struct {
	shared.BaseModel
	Name        string  `json:"name"`        // Nama pekerjaan (misalnya, Pembangunan dinding)
	Description string  `json:"description"` // Nama pekerjaan (misalnya, Pembangunan dinding)
	UnitName    string  `json:"unit_name"`   // Nama unit pekerjaan (misalnya, meter persegi)
	UnitPrice   float64 `json:"unit_price"`  // Harga per unit pekerjaan
}

type WorkReportCheck struct {
	shared.BaseModel
	EmployeeID     *string       `json:"employee_id"`
	Employee       EmployeeModel `gorm:"foreignKey:EmployeeID"`
	WorkReportID   string        `json:"work_report_id"`
	AmountReported float64       `json:"amount_reported"`
	AmountApproved float64       `json:"amount_approved"`
	AmountRejected float64       `json:"amount_rejected"`
	Remarks        string        `json:"remarks"`
	CompanyID      string        `json:"company_id"`
	Company        CompanyModel  `gorm:"foreignKey:CompanyID"`
	CheckedAt      time.Time     `json:"checked_at"`
	Pictures       []FileModel   `json:"pictures" gorm:"-"`
}
