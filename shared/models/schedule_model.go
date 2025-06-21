package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ScheduleModel struct {
	shared.BaseModel
	Name            string              `json:"name" gorm:"type:varchar(255)"`
	Description     string              `json:"description" gorm:"type:text"`
	Code            string              `json:"code" gorm:"type:varchar(255)"`
	ScheduleType    string              `json:"schedule_type" gorm:"type:varchar(255)"`
	WeekDay         *string             `json:"week_day" gorm:"type:varchar(255)"`
	StartDate       *time.Time          `json:"start_date" gorm:"type:DATE" sql:"TYPE:DATE"`
	Date            string              `json:"date" sql:"event_date"`
	EndDate         *time.Time          `json:"end_date" gorm:"type:DATE" sql:"TYPE:DATE"`
	StartTime       *time.Time          `json:"start_time" gorm:"type:TIME"`
	EndTime         *time.Time          `json:"end_time" gorm:"type:TIME"`
	Employees       []EmployeeModel     `json:"-" gorm:"many2many:schedule_employees;constraint:OnDelete:CASCADE;"`
	Organizations   []OrganizationModel `json:"-" gorm:"many2many:schedule_organizations;constraint:OnDelete:CASCADE;"`
	Branchs         []BranchModel       `json:"-" gorm:"many2many:schedule_branchs;constraint:OnDelete:CASCADE;"`
	Sunday          bool                `json:"sunday"`
	Monday          bool                `json:"monday"`
	Tuesday         bool                `json:"tuesday"`
	Wednesday       bool                `json:"wednesday"`
	Thursday        bool                `json:"thursday"`
	Friday          bool                `json:"friday"`
	Saturday        bool                `json:"saturday"`
	EmployeeIDs     []string            `json:"employee_ids" gorm:"type:varchar(255)"`
	OrganizationIDs []string            `json:"organization_ids" gorm:"type:varchar(255)"`
	BranchIDs       []string            `json:"branch_ids" gorm:"type:varchar(255)"`
	CompanyID       *string             `json:"company_id" gorm:"type:char(36);not null"`
	Company         *CompanyModel       `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE"`
	UserID          *string             `json:"user_id" gorm:"type:char(36)"`
	User            *UserModel          `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	ExternalID      string              `json:"external_id" gorm:"type:varchar(255)"`
	Source          string              `json:"string" gorm:"type:varchar(255)"`
	Data            string              `json:"data" gorm:"type:JSON"`
	IsPolicyChecked bool                `json:"is_policy_checked"`
	WorkShiftID     *string             `json:"work_shift_id" gorm:"type:varchar(255)"`
	WorkShift       WorkShiftModel      `gorm:"foreignKey:WorkShiftID"`
	EffectiveDate   *time.Time          `json:"effective_date" gorm:"type:DATE" sql:"TYPE:DATE"`
	EffectiveUntil  *time.Time          `json:"effective_until" gorm:"type:DATE" sql:"TYPE:DATE"`
	RepeatEvery     int64               `json:"repeat_every" gorm:"-"` // in a day
	RepeatPause     int64               `json:"repeat_pause" gorm:"-"` // count repeat every and pause after repeat = pause or multiple
	RepeatGap       int64               `json:"repeat_gap" gorm:"-"`   // pause gap in a day and start again if loop current date > repeat gap
	RepeatUntil     *time.Time          `json:"repeat_until" gorm:"-"` //
}

func (ScheduleModel) TableName() string {
	return "schedules"
}

func (s *ScheduleModel) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	return nil
}
