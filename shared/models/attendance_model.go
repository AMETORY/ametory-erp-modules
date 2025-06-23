package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AttendanceModel struct {
	shared.BaseModel
	ClockIn                    time.Time               `json:"clock_in,omitempty"`
	ClockOut                   *time.Time              `json:"clock_out,omitempty"`
	ClockInNotes               string                  `json:"clock_in_notes,omitempty"`
	ClockOutNotes              string                  `json:"clock_out_notes,omitempty"`
	ClockInPicture             string                  `json:"clock_in_picture,omitempty"`
	ClockOutPicture            string                  `json:"clock_out_picture,omitempty"`
	ClockInLat                 *float64                `json:"clock_in_lat,omitempty" gorm:"type:DECIMAL(10,8)"`
	ClockInLng                 *float64                `json:"clock_in_lng,omitempty" gorm:"type:DECIMAL(11,8)"`
	ClockOutLat                *float64                `json:"clock_out_lat,omitempty" gorm:"type:DECIMAL(10,8)"`
	ClockOutLng                *float64                `json:"clock_out_lng,omitempty" gorm:"type:DECIMAL(11,8)"`
	EmployeeID                 *string                 `json:"employee_id,omitempty"`
	Employee                   *EmployeeModel          `gorm:"foreignKey:EmployeeID;constraint:OnDelete:CASCADE" json:"employee,omitempty"`
	BreakStart                 *time.Time              `json:"break_start,omitempty"`
	BreakEnd                   *time.Time              `json:"break_end,omitempty"`
	Overtime                   *time.Duration          `json:"overtime,omitempty"`
	LateIn                     *time.Duration          `json:"late_in,omitempty"`
	WorkingDuration            *time.Duration          `json:"working_duration,omitempty"`
	CompanyID                  *string                 `json:"company_id,omitempty" gorm:"not null"`
	Company                    *CompanyModel           `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE" json:"company,omitempty"`
	Activities                 []EmployeeActivityModel `json:"activities,omitempty" gorm:"foreignKey:AttendanceID;constraint:OnDelete:CASCADE"`
	WorkReports                []WorkReport            `json:"work_reports,omitempty" gorm:"foreignKey:AttendanceID;constraint:OnDelete:CASCADE" `
	Status                     string                  `json:"status,omitempty" gorm:"default:'ACTIVE'"`
	Remarks                    string                  `json:"remarks,omitempty" gorm:"type:TEXT"`
	ClockOutRemarks            string                  `json:"clock_out_remarks,omitempty" gorm:"type:TEXT"`
	AttendancePolicyID         *string                 `json:"attendance_policy_id,omitempty"`
	AttendancePolicy           *AttendancePolicy       `gorm:"foreignKey:AttendancePolicyID" json:"attendance_policy,omitempty"`
	ClockOutAttendancePolicyID *string                 `json:"clock_out_attendance_policy_id,omitempty"`
	ClockOutAttendancePolicy   *AttendancePolicy       `gorm:"foreignKey:ClockOutAttendancePolicyID" json:"clock_out_attendance_policy,omitempty"`
	ScheduleID                 *string                 `json:"schedule_id,omitempty"`
	Schedule                   *ScheduleModel          `gorm:"foreignKey:ScheduleID" json:"schedule,omitempty"`
	BranchID                   *string                 `json:"branch_id,omitempty"`
	Branch                     *BranchModel            `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	OrganizationID             *string                 `json:"organization_id,omitempty"`
	Organization               *OrganizationModel      `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	WorkShiftID                *string                 `json:"work_shift_id,omitempty"`
	WorkShift                  *WorkShiftModel         `gorm:"foreignKey:WorkShiftID" json:"work_shift,omitempty"`
	Timezone                   string                  `gorm:"-"`
	// AttendanceBulkImportID     *string              `json:"attendance_bulk_import_id"`
	// AttendanceBulkImport       AttendanceBulkImport `gorm:"foreignKey:AttendanceBulkImportID"`
	// AttendanceImportItemID     *string              `json:"attendance_import_item_id"`
	// AttendanceImportItem       AttendanceImportItem `gorm:"foreignKey:AttendanceImportItemID"`
}

func (a AttendanceModel) TableName() string {
	return "attendances"
}

func (a *AttendanceModel) BeforeCreate(tx *gorm.DB) (err error) {
	if a.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}

type AttendanceCheckInput struct {
	Now               time.Time  `json:"now"`
	Lat               *float64   `json:"lat"`
	Lng               *float64   `json:"lng"`
	IsFaceDetected    bool       `json:"is_face_detected"`
	IsClockIn         bool       `json:"is_clock_in"`
	ScheduledClockIn  time.Time  `json:"scheduled_clock_in"`
	ScheduledClockOut time.Time  `json:"scheduled_clock_out"`
	AttendanceID      *string    `json:"attendance_id"`
	EmployeeID        *string    `json:"employee_id"`
	ScheduleID        *string    `json:"schedule_id"`
	Notes             string     `json:"notes"`
	File              *FileModel `json:"file"`
}
