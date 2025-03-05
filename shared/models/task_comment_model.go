package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TaskCommentModel struct {
	shared.BaseModel
	TaskID      string       `gorm:"type:char(36);index" json:"task_id,omitempty"`
	MemberID    *string      `gorm:"type:char(36);index" json:"member_id,omitempty"`
	Member      *MemberModel `gorm:"foreignKey:MemberID;constraint:OnDelete:CASCADE;" json:"member,omitempty"`
	Comment     string       `json:"comment,omitempty"`
	Status      string       `json:"status,omitempty"`
	PublishedAt *time.Time   `json:"published_at,omitempty"`
}

func (TaskCommentModel) TableName() string {
	return "task_comments"
}

func (m *TaskCommentModel) BeforeCreate(tx *gorm.DB) error {

	if m.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return nil
}
