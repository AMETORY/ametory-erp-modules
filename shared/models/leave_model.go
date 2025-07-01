package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LeaveModel struct {
	shared.BaseModel
	Name              string         `json:"name"`
	RequestType       string         `json:"request_type" gorm:"default:'FULL_DAY'"`
	LeaveCategoryID   *string        `json:"leave_category_id"`
	LeaveCategory     LeaveCategory  `json:"leave_category" gorm:"foreignKey:LeaveCategoryID"`
	StartDate         *time.Time     `json:"start_date" gorm:"type:DATE" sql:"TYPE:DATE"`
	EndDate           *time.Time     `json:"end_date" gorm:"type:DATE" sql:"TYPE:DATE"`
	StartTime         *TimeOnly      `json:"start_time" gorm:"type:TIME"`
	EndTime           *TimeOnly      `json:"end_time" gorm:"type:TIME"`
	EmployeeID        *string        `json:"employee_id"`
	Employee          *EmployeeModel `json:"employee" gorm:"foreignKey:EmployeeID"`
	Description       string         `json:"description" gorm:"type:TEXT"`
	Status            string         `json:"status" gorm:"default:'DRAFT'"`
	Remarks           string         `json:"remarks" gorm:"type:TEXT"`
	Attachment        *string        `json:"attachment" gorm:"type:TEXT"`
	ApproverID        *string        `json:"approver_id"`
	Approver          *EmployeeModel `json:"approver" gorm:"foreignKey:ApproverID"`
	ApprovalDate      *time.Time     `json:"approval_date"`
	ApprovalByAdminID *string        `json:"approval_by_admin_id"`
	ApprovalByAdmin   *UserModel     `json:"approval_by_admin" gorm:"foreignKey:ApprovalByAdminID"`
	CompanyID         *string        `json:"company_id" gorm:"not null"`
	Company           *CompanyModel  `gorm:"foreignKey:CompanyID"`
	ScheduleID        *string        `json:"schedule_id"`
	Schedule          *ScheduleModel `json:"schedule" gorm:"foreignKey:ScheduleID"`
	Files             []FileModel    `json:"files" gorm:"-"`
}

func (p *LeaveModel) TableName() string {
	return "leaves"
}

func (p *LeaveModel) BeforeCreate(tx *gorm.DB) (err error) {

	if p.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}

type LeaveCategory struct {
	shared.BaseModel
	Name            string        `gorm:"uniqueIndex;type:VARCHAR(255)" json:"name,omitempty"`
	Description     string        `json:"description,omitempty"`
	Absent          bool          `gorm:"default:false; NOT NULL" json:"absent,omitempty"`
	Sick            bool          `gorm:"default:false; NOT NULL" json:"sick,omitempty"`
	CompanyID       *string       `json:"company_id,omitempty"`
	Company         *CompanyModel `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	IsYearlyQuota   bool          `json:"is_yearly_quota,omitempty" gorm:"not null"`
	IsShiftSchedule bool          `json:"is_shift_schedule,omitempty" gorm:"not null"`
}

func (p *LeaveCategory) TableName() string {
	return "leave_categories"
}

func (p *LeaveCategory) BeforeCreate(tx *gorm.DB) (err error) {

	if p.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}
