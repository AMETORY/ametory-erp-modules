package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EmployeeActivityModel struct {
	shared.BaseModel
	Name              string                `json:"name"`
	ActivityType      string                `json:"activity_type" gorm:"default:'TO_DO'"`
	StartDate         *time.Time            `json:"start_date" gorm:"type:DATE" sql:"TYPE:DATE"`
	EndDate           *time.Time            `json:"end_date" gorm:"type:DATE" sql:"TYPE:DATE"`
	StartTime         *time.Time            `json:"start_time" gorm:"type:TIME"`
	EndTime           *time.Time            `json:"end_time" gorm:"type:TIME"`
	EmployeeID        string                `json:"employee_id"`
	Employee          EmployeeModel         `json:"employee" gorm:"foreignKey:EmployeeID"`
	AssignedToID      *string               `json:"assigned_to_id"`
	AssignedTo        *EmployeeModel        `json:"assigned_to" gorm:"foreignKey:AssignedToID"`
	Description       string                `json:"description" gorm:"type:TEXT"`
	Status            string                `json:"status" gorm:"default:'DRAFT'"`
	Remarks           string                `json:"remarks" gorm:"type:TEXT"`
	Attachment        *string               `json:"attachment" gorm:"type:TEXT"`
	CompanyID         string                `json:"company_id" gorm:"not null"`
	Company           CompanyModel          `gorm:"foreignKey:CompanyID"`
	Files             []FileModel           `json:"files" gorm:"constraint:OnDelete:CASCADE"`
	Lat               *float64              `json:"lat" gorm:"type:DECIMAL(10,8)"`
	Lng               *float64              `json:"lng" gorm:"type:DECIMAL(11,8)"`
	Location          *string               `json:"location" gorm:"type:TEXT"`
	AttendanceID      *string               `json:"attendance_id" gorm:"type:char(36)"`
	IsAssignment      bool                  `gorm:"-"`
	ApproverID        *string               `json:"approver_id"`
	Approver          EmployeeModel         `json:"approver" gorm:"foreignKey:ApproverID"`
	ApprovalDate      *time.Time            `json:"approval_date"`
	ApprovalRemarks   *string               `json:"approval_remarks"`
	ApprovalStatus    string                `json:"approval_status" gorm:"default:'PENDING'"`
	IsNeedApproval    bool                  `json:"is_need_approval" gorm:"-"`
	OvertimeRequestID *string               `json:"overtime_request_id"`
	OvertimeRequest   EmployeeOvertimeModel `gorm:"foreignKey:OvertimeRequestID"`
	StartTimePicture  string                `json:"start_time_picture"`
	EndTimePicture    string                `json:"end_time_picture"`
	VisitOutLat       *float64              `json:"visit_out_lat" gorm:"type:DECIMAL(10,8)"`
	VisitOutLng       *float64              `json:"visit_out_lng" gorm:"type:DECIMAL(11,8)"`
	VisitOutLocation  *string               `json:"visit_out_location" gorm:"type:TEXT"`
}

func (e EmployeeActivityModel) TableName() string {
	return "employee_activities"
}

func (e *EmployeeActivityModel) BeforeCreate(tx *gorm.DB) error {
	e.ID = uuid.New().String()
	return nil
}
