package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AttendanceModel struct {
	shared.BaseModel
	ClockIn                    time.Time               `json:"clock_in"`
	ClockOut                   *time.Time              `json:"clock_out"`
	ClockInNotes               string                  `json:"clock_in_notes"`
	ClockOutNotes              string                  `json:"clock_out_notes"`
	ClockInPicture             string                  `json:"clock_in_picture"`
	ClockOutPicture            string                  `json:"clock_out_picture"`
	ClockInLat                 *float64                `json:"clock_in_lat" gorm:"type:DECIMAL(10,8)"`
	ClockInLng                 *float64                `json:"clock_in_lng" gorm:"type:DECIMAL(11,8)"`
	ClockOutLat                *float64                `json:"clock_out_lat" gorm:"type:DECIMAL(10,8)"`
	ClockOutLng                *float64                `json:"clock_out_lng" gorm:"type:DECIMAL(11,8)"`
	EmployeeID                 *string                 `json:"employee_id"`
	Employee                   EmployeeModel           `gorm:"foreignKey:EmployeeID;constraint:OnDelete:CASCADE"`
	BreakStart                 *time.Time              `json:"break_start" `
	BreakEnd                   *time.Time              `json:"break_end" `
	Overtime                   *time.Duration          `json:"overtime" `
	LateIn                     *time.Duration          `json:"late_in" `
	WorkingDuration            *time.Duration          `json:"working_duration" `
	CompanyID                  string                  `json:"company_id" gorm:"not null"`
	Company                    CompanyModel            `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE"`
	Activities                 []EmployeeActivityModel `json:"activities" gorm:"constraint:OnDelete:CASCADE"`
	WorkReports                []WorkReport            `json:"work_reports" gorm:"constraint:OnDelete:CASCADE"`
	Status                     string                  `json:"status" gorm:"default:'ACTIVE'"`
	Remarks                    string                  `json:"remarks" gorm:"type:TEXT"`
	ClockOutRemarks            string                  `json:"clock_out_remarks" gorm:"type:TEXT"`
	AttendancePolicyID         *string                 `json:"attendance_policy_id"`
	AttendancePolicy           AttendancePolicy        `gorm:"foreignKey:AttendancePolicyID"`
	ClockOutAttendancePolicyID *string                 `json:"clock_out_attendance_policy_id"`
	ClockOutAttendancePolicy   AttendancePolicy        `gorm:"foreignKey:ClockOutAttendancePolicyID"`
	ScheduleID                 *string                 `json:"schedule_id"`
	Schedule                   ScheduleModel           `gorm:"foreignKey:ScheduleID"`
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
