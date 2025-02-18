package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
)

type ContentCommentModel struct {
	shared.BaseModel
	RefID       string     `gorm:"type:char(36);index" json:"ref_id,omitempty"`
	RefType     string     `gorm:"type:varchar(255);index" json:"ref_type,omitempty"`
	UserID      *string    `gorm:"type:char(36);index" json:"user_id,omitempty"`
	User        *UserModel `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;" json:"user,omitempty"`
	Comment     string     `json:"comment,omitempty"`
	FullName    *string    `json:"full_name,omitempty"`
	Email       *string    `json:"email,omitempty"`
	Status      string     `json:"status,omitempty"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
}
