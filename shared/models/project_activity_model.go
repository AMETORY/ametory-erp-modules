package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProjectActivityModel struct {
	shared.BaseModel
	ProjectID    string       `gorm:"type:char(36);not null" json:"project_id"`
	Project      ProjectModel `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE;" json:"project"`
	MemberID     string       `gorm:"type:char(36);not null" json:"member_id"`
	Member       MemberModel  `gorm:"foreignKey:MemberID;constraint:OnDelete:CASCADE;" json:"member"`
	ActivityType string       `gorm:"not null" json:"activity_type"`
	Notes        *string      `gorm:"type:text" json:"notes,omitempty"`
	ColumnID     *string      `gorm:"type:char(36)" json:"column_id,omitempty"`
	Column       *ColumnModel `gorm:"foreignKey:ColumnID;constraint:OnDelete:CASCADE;" json:"column,omitempty"`
	TaskID       *string      `gorm:"type:char(36)" json:"task_id,omitempty"`
	Task         *TaskModel   `gorm:"foreignKey:TaskID;constraint:OnDelete:CASCADE;" json:"task,omitempty"`
	ActivityDate *time.Time   `json:"activity_date,omitempty"`
}

func (m *ProjectActivityModel) TableName() string {
	return "project_activities"
}

func (m *ProjectActivityModel) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	if m.ActivityDate == nil {
		now := time.Now()
		m.ActivityDate = &now
	}
	return nil
}
