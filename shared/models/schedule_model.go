package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type ScheduleModel struct {
	shared.BaseModel
	Name            string              `json:"name,omitempty" gorm:"type:varchar(255)"`
	Description     string              `json:"description,omitempty" gorm:"type:text"`
	Code            string              `json:"code,omitempty" gorm:"type:varchar(255)"`
	ScheduleType    string              `json:"schedule_type,omitempty" gorm:"type:varchar(255)"`
	WeekDay         *string             `json:"week_day,omitempty" gorm:"type:varchar(255)"`
	StartDate       *time.Time          `json:"start_date,omitempty" gorm:"type:DATE" sql:"TYPE:DATE"`
	Date            string              `json:"date,omitempty" sql:"event_date"`
	EndDate         *time.Time          `json:"end_date,omitempty" gorm:"type:DATE" sql:"TYPE:DATE"`
	StartTime       *string             `json:"start_time,omitempty"`
	EndTime         *string             `json:"end_time,omitempty"`
	RepeatType      string              `json:"repeat_type,omitempty"` // "ONCE", "DAILY", "WEEKLY"
	RepeatDays      pq.StringArray      `gorm:"type:text[]" json:"repeat_days,omitempty"`
	Employees       []EmployeeModel     `json:"employees,omitempty" gorm:"many2many:schedule_employees;constraint:OnDelete:CASCADE;"`
	Organizations   []OrganizationModel `json:"organizations,omitempty" gorm:"many2many:schedule_organizations;constraint:OnDelete:CASCADE;"`
	Branches        []BranchModel       `json:"branches,omitempty" gorm:"many2many:schedule_branches;constraint:OnDelete:CASCADE;"`
	IsActive        bool                `json:"is_active,omitempty"`
	CompanyID       *string             `json:"company_id,omitempty" gorm:"type:char(36);not null"`
	Company         *CompanyModel       `json:"company,omitempty" gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE"`
	UserID          *string             `json:"user_id,omitempty" gorm:"type:char(36)"`
	User            *UserModel          `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	ExternalID      string              `json:"external_id,omitempty" gorm:"type:varchar(255)"`
	Source          string              `json:"source,omitempty" gorm:"type:varchar(255)"`
	Data            string              `json:"data,omitempty" gorm:"type:JSON"`
	IsPolicyChecked bool                `json:"is_policy_checked,omitempty"`
	WorkShiftID     *string             `json:"work_shift_id,omitempty" gorm:"type:varchar(255)"`
	WorkShift       *WorkShiftModel     `json:"work_shift,omitempty" gorm:"foreignKey:WorkShiftID"`
	EffectiveDate   *time.Time          `json:"effective_date,omitempty" gorm:"type:DATE" sql:"TYPE:DATE"`
	EffectiveUntil  *time.Time          `json:"effective_until,omitempty" gorm:"type:DATE" sql:"TYPE:DATE"`
	RepeatEvery     int64               `json:"repeat_every,omitempty" gorm:"-"` // in a day
	RepeatPause     int64               `json:"repeat_pause,omitempty" gorm:"-"` // count repeat every and pause after repeat = pause or multiple
	RepeatGap       int64               `json:"repeat_gap,omitempty" gorm:"-"`   // pause gap in a day and start again if loop current date > repeat gap
	RepeatUntil     *time.Time          `json:"repeat_until,omitempty" gorm:"-"`
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
