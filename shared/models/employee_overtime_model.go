package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EmployeeOvertimeModel struct {
	shared.BaseModel
	CompanyID                *string          `json:"company_id" gorm:"not null"`
	Company                  *CompanyModel    `gorm:"foreignKey:CompanyID"`
	EmployeeID               *string          `json:"employee_id"`
	Employee                 *EmployeeModel   `gorm:"foreignKey:EmployeeID" json:"employee"`
	ApproverID               *string          `json:"approver_id"`
	Approver                 *EmployeeModel   `gorm:"foreignKey:ApproverID" json:"approver"`
	ReviewerID               *string          `json:"reviewer_id"`
	Reviewer                 *EmployeeModel   `gorm:"foreignKey:ReviewerID" json:"reviewer"`
	StartTimeRequest         time.Time        `json:"start_time_request"`
	EndTimeRequest           time.Time        `json:"end_time_request"`
	Reason                   string           `json:"reason"`
	Status                   string           `json:"status" gorm:"default:'PENDING'"` // 'PENDING','APPROVED', 'REJECTED','FINISHED', 'REVIEWED'
	DateRequested            time.Time        `json:"date_requested"`
	DateApprovedOrRejected   *time.Time       `json:"date_approved_or_rejected"`
	ClockIn                  *time.Time       `json:"clock_in"`
	ClockOut                 *time.Time       `json:"clock_out"`
	ClockInNotes             string           `json:"clock_in_notes"`
	ClockOutNotes            string           `json:"clock_out_notes"`
	ClockInPicture           string           `json:"clock_in_picture"`
	ClockOutPicture          string           `json:"clock_out_picture"`
	ClockInLat               *float64         `json:"clock_in_lat" gorm:"type:DECIMAL(10,8)"`
	ClockInLng               *float64         `json:"clock_in_lng" gorm:"type:DECIMAL(11,8)"`
	ClockOutLat              *float64         `json:"clock_out_lat" gorm:"type:DECIMAL(10,8)"`
	ClockOutLng              *float64         `json:"clock_out_lng" gorm:"type:DECIMAL(11,8)"`
	Remarks                  string           `json:"remarks"`
	OvertimeDurationApproved *int             `json:"overtime_duration_approved"`
	AttendanceID             *string          `json:"attendance_id"`
	Attendance               *AttendanceModel `gorm:"foreignKey:AttendanceID" json:"attendance"`
	EmployeeApproverID       *string          `json:"employee_approver_id"`
	EmployeeApprover         *EmployeeModel   `gorm:"foreignKey:EmployeeApproverID"`
	ApprovalByAdminID        *string          `json:"approval_by_admin_id"`
	ApprovalByAdmin          *UserModel       `json:"approval_by_admin" gorm:"foreignKey:ApprovalByAdminID"`
}

func (EmployeeOvertimeModel) TableName() string {
	return "employee_overtimes"
}

func (e *EmployeeOvertimeModel) BeforeCreate(tx *gorm.DB) (err error) {

	if e.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return nil
}
