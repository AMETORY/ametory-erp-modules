package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AttendanceStatus string

const (
	Reject  AttendanceStatus = "REJECT"
	Pending AttendanceStatus = "PENDING"
	Active  AttendanceStatus = "ACTIVE"
)

type Remarks string

const (
	FaceProblem             Remarks = "FACE"
	LocationProblem         Remarks = "LOCATION"
	LocationNotFoundProblem Remarks = "LOCATION_NOT_FOUND"
	LocationDistanceProblem Remarks = "LOCATION_DISTANCE"
	ScheduleProblem         Remarks = "SCHEDULE"
	EarlyInProblem          Remarks = "EARLY_IN"
	LateInProblem           Remarks = "LATE_IN"
	EarlyOutProblem         Remarks = "EARLY_OUT"
	LateOutProblem          Remarks = "LATE_OUT"
	CustomProblem           Remarks = "CUSTOM"
	BranchProblem           Remarks = "BRANCH"
	OrganizationProblem     Remarks = "ORGANIZATION"
	NoPolicyProblem         Remarks = "NO_POLICY"
)

type WorkingType string

const (
	FullTime  WorkingType = "FULL_TIME"
	PartTime  WorkingType = "PART_TIME"
	Freelance WorkingType = "FREELANCE"
	Flexible  WorkingType = "FLEXIBLE"
	ShiftType WorkingType = "SHIFT"
	Seasonal  WorkingType = "SEASONAL"
)

type Shift string

const (
	Morning Shift = "MORNING"
	Evening Shift = "EVENING"
	Night   Shift = "NIGHT"
)

type AttendancePolicyReq struct {
	Time           time.Time
	Employee       EmployeeModel
	Lat            *float64
	Lng            *float64
	Schedules      []ScheduleModel
	IsFaceDetected bool
	IsClockOut     bool
}

type CustomCondition func(attendanceReq AttendancePolicyReq, policy AttendancePolicy) bool

type AttendancePolicy struct {
	shared.BaseModel
	PolicyName              string             `json:"policy_name"`
	Description             string             `json:"description"`
	Priority                int                `json:"priority"`
	CompanyID               *string            `json:"company_id"`
	Company                 *CompanyModel      `gorm:"foreignKey:CompanyID"`
	LocationEnabled         bool               `json:"location_enabled"`
	Lat                     *float64           `json:"lat" gorm:"type:DECIMAL(10,8)"`
	Lng                     *float64           `json:"lng" gorm:"type:DECIMAL(11,8)"`
	MaxAttendanceRadius     *float64           `json:"max_attendance_radius"`
	ScheduledCheck          bool               `json:"scheduled_check"`
	OnPolicyFailure         AttendanceStatus   `json:"on_policy_failure" gorm:"default:'REJECT'"`
	OnLocationFailure       AttendanceStatus   `json:"on_location_failure" gorm:"default:'REJECT'"`
	OnScheduleFailure       AttendanceStatus   `json:"on_schedule_failure" gorm:"default:'REJECT'"`
	OnClockInFailure        AttendanceStatus   `json:"on_clock_in_failure" gorm:"default:'REJECT'"`
	OnEarlyInFailure        AttendanceStatus   `json:"on_early_in_failure" gorm:"default:'REJECT'"`
	OnClockOutFailure       AttendanceStatus   `json:"on_clock_out_failure" gorm:"default:'REJECT'"`
	OnEarlyOutFailure       AttendanceStatus   `json:"on_early_out_failure" gorm:"default:'REJECT'"`
	OnFaceNotDetected       AttendanceStatus   `json:"on_face_not_detected" gorm:"default:'REJECT'"`
	CustomConditions        []CustomCondition  `gorm:"-"`
	IsClockOut              bool               `gorm:"-"`
	CustomConditionData     string             `gorm:"type:JSON" json:"custom_condition_data"`
	Remarks                 Remarks            `gorm:"-"`
	IsActive                bool               `json:"is_active"`
	BranchID                *string            `json:"branch_id"`
	Branch                  *BranchModel       `gorm:"foreignKey:BranchID"`
	OrganizationID          *string            `json:"organization_id"`
	Organization            *OrganizationModel `gorm:"foreignKey:OrganizationID"`
	WorkShiftID             *string            `json:"work_shift_id"`
	WorkShift               *WorkShiftModel    `gorm:"foreignKey:WorkShiftID"`
	EarlyInToleranceInTime  time.Duration      `json:"early_in_tolerance" gorm:"default:15"`
	LateInToleranceInTime   time.Duration      `json:"late_in_tolerance" gorm:"default:15"`
	EarlyOutToleranceInTime time.Duration      `json:"early_out_tolerance" gorm:"default:15"`
	LateOutToleranceInTime  time.Duration      `json:"late_out_tolerance" gorm:"default:60"`
	ScheduleID              *string            `gorm:"-"`
	Schedule                *ScheduleModel     `gorm:"-"`
}

func (u *AttendancePolicy) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}

	return
}

func (AttendancePolicy) TableName() string {
	return "attendance_policies"
}
