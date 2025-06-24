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
	Empty                   Remarks = ""
)

var ProblemDescriptions = map[Remarks]map[string]string{
	FaceProblem: {
		"EN": "Face not detected",
		"ID": "Wajah tidak terdeteksi",
	},
	LocationProblem: {
		"EN": "Location not detected",
		"ID": "Lokasi tidak terdeteksi",
	},
	LocationNotFoundProblem: {
		"EN": "Location not found",
		"ID": "Lokasi tidak ditemukan",
	},
	LocationDistanceProblem: {
		"EN": "Distance from location is too far",
		"ID": "Jarak dari lokasi terlalu jauh",
	},
	ScheduleProblem: {
		"EN": "Schedule not found",
		"ID": "Jadwal tidak ditemukan",
	},
	EarlyInProblem: {
		"EN": "Clock in too early",
		"ID": "Clock In terlalu awal",
	},
	LateInProblem: {
		"EN": "Clock in too late",
		"ID": "Clock In terlambat",
	},
	EarlyOutProblem: {
		"EN": "Clock out too early",
		"ID": "Clock Out terlalu awal",
	},
	LateOutProblem: {
		"EN": "Clock out too late",
		"ID": "Clock Out terlambat",
	},
	CustomProblem: {
		"EN": "Custom rules failed",
		"ID": "Aturan khusus gagal",
	},
	BranchProblem: {
		"EN": "Branch not found",
		"ID": "Cabang tidak ditemukan",
	},
	OrganizationProblem: {
		"EN": "Organization not found",
		"ID": "Organisasi tidak ditemukan",
	},
	NoPolicyProblem: {
		"EN": "No attendance policy found",
		"ID": "Kebijakan kehadiran tidak ditemukan",
	},
	Empty: {
		"EN": "Unknown error",
		"ID": "Error tidak diketahui",
	},
}

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
	PolicyName              string             `json:"policy_name,omitempty"`
	Description             string             `json:"description,omitempty"`
	Priority                int                `json:"priority,omitempty"`
	CompanyID               *string            `json:"company_id,omitempty"`
	Company                 *CompanyModel      `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	LocationEnabled         bool               `json:"location_enabled,omitempty"`
	Lat                     *float64           `json:"lat,omitempty" gorm:"type:DECIMAL(10,8)"`
	Lng                     *float64           `json:"lng,omitempty" gorm:"type:DECIMAL(11,8)"`
	MaxAttendanceRadius     *float64           `json:"max_attendance_radius,omitempty"`
	ScheduledCheck          bool               `json:"scheduled_check,omitempty"`
	OnPolicyFailure         AttendanceStatus   `json:"on_policy_failure,omitempty" gorm:"default:'REJECT'"`
	OnLocationFailure       AttendanceStatus   `json:"on_location_failure,omitempty" gorm:"default:'REJECT'"`
	OnScheduleFailure       AttendanceStatus   `json:"on_schedule_failure,omitempty" gorm:"default:'REJECT'"`
	OnClockInFailure        AttendanceStatus   `json:"on_clock_in_failure,omitempty" gorm:"default:'REJECT'"`
	OnEarlyInFailure        AttendanceStatus   `json:"on_early_in_failure,omitempty" gorm:"default:'REJECT'"`
	OnClockOutFailure       AttendanceStatus   `json:"on_clock_out_failure,omitempty" gorm:"default:'REJECT'"`
	OnEarlyOutFailure       AttendanceStatus   `json:"on_early_out_failure,omitempty" gorm:"default:'REJECT'"`
	OnFaceNotDetected       AttendanceStatus   `json:"on_face_not_detected,omitempty" gorm:"default:'REJECT'"`
	CustomConditions        []CustomCondition  `gorm:"-"`
	IsClockOut              bool               `gorm:"-"`
	CustomConditionData     string             `gorm:"type:JSON" json:"custom_condition_data,omitempty"`
	Remarks                 Remarks            `gorm:"-"`
	IsActive                bool               `json:"is_active,omitempty"`
	BranchID                *string            `json:"branch_id,omitempty"`
	Branch                  *BranchModel       `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	OrganizationID          *string            `json:"organization_id,omitempty"`
	Organization            *OrganizationModel `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	WorkShiftID             *string            `json:"work_shift_id,omitempty"`
	WorkShift               *WorkShiftModel    `gorm:"foreignKey:WorkShiftID" json:"work_shift,omitempty"`
	EarlyInToleranceInTime  time.Duration      `json:"early_in_tolerance,omitempty" gorm:"default:15"`
	LateInToleranceInTime   time.Duration      `json:"late_in_tolerance,omitempty" gorm:"default:15"`
	EarlyOutToleranceInTime time.Duration      `json:"early_out_tolerance,omitempty" gorm:"default:15"`
	LateOutToleranceInTime  time.Duration      `json:"late_out_tolerance,omitempty" gorm:"default:60"`
	EffectiveDate           *time.Time         `json:"effective_date,omitempty" gorm:"type:DATE" sql:"TYPE:DATE"`
	EffectiveUntil          *time.Time         `json:"effective_until,omitempty" gorm:"type:DATE" sql:"TYPE:DATE"`
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
