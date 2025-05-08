package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserActivityType string

const (
	UserActivityLogin      UserActivityType = "LOGIN"
	UserActivityLogout     UserActivityType = "LOGOUT"
	UserActivityBreak      UserActivityType = "BREAK"
	UserActivityBreakOff   UserActivityType = "BREAK_OFF"
	UserActivityMeeting    UserActivityType = "MEETING"
	UserActivityTraining   UserActivityType = "TRAINING"
	UserActivityOnline     UserActivityType = "ONLINE"
	UserActivityOffline    UserActivityType = "OFFLINE"
	UserActivityClockIn    UserActivityType = "CLOCK_IN"
	UserActivityClockOut   UserActivityType = "CLOCK_OUT"
	UserActivityCheckPoint UserActivityType = "CHECK_POINT"
	UserActivityAttend     UserActivityType = "ATTEND"
	UserActivityWorkOn     UserActivityType = "WORK_ON"
	UserActivitySales      UserActivityType = "SALES"
	UserActivityPurchase   UserActivityType = "PURCHASE"
	UserActivityApprove    UserActivityType = "APPROVE"
	UserActivityDecline    UserActivityType = "DECLINE"
)

type UserActivityModel struct {
	shared.BaseModel
	UserID            string           `gorm:"not null;index" json:"user_id"`
	User              *UserModel       `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;" json:"-"`
	ActivityType      UserActivityType `gorm:"not null" json:"activity_type"`
	Latitude          *float64         `json:"latitude" gorm:"type:decimal(10,8)"`
	Longitude         *float64         `json:"longitude" gorm:"type:decimal(11,8)"`
	LocationID        *string          `gorm:"index" json:"location_id,omitempty"`
	Duration          *time.Duration   `json:"duration,omitempty"`
	StartedAt         *time.Time       `json:"started_at,omitempty"`
	FinishedAt        *time.Time       `json:"finished_at,omitempty"`
	Notes             *string          `gorm:"type:text" json:"notes,omitempty"`
	RefID             *string          `gorm:"type:char(36)" json:"ref_id,omitempty"`
	RefType           *string          `gorm:"type:varchar(255)" json:"ref_type,omitempty"`
	Files             []FileModel      `gorm:"-" json:"files,omitempty"`
	FinishedLatitude  *float64         `json:"finished_latitude" gorm:"type:decimal(10,8)"`
	FinishedLongitude *float64         `json:"finished_longitude" gorm:"type:decimal(11,8)"`
	FinishedNotes     *string          `gorm:"type:text" json:"finished_notes,omitempty"`
}

func (m *UserActivityModel) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	return nil
}

func (m *UserActivityModel) AfterFind(tx *gorm.DB) error {
	var files []FileModel
	tx.Where("ref_id = ? AND ref_type in (?)", m.ID, []string{"user_activity", "clock_in", "clock_out", "check_point"}).Find(&files)
	m.Files = files
	return nil
}
func (m *UserActivityModel) TableName() string {
	return "user_activities"
}
