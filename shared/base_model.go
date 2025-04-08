package shared

import (
	"time"

	"gorm.io/gorm"
)

type BaseModel struct {
	ID        string         `gorm:"type:char(36);primary_key" json:"id,omitempty"`
	CreatedAt *time.Time     `json:"-"`
	UpdatedAt *time.Time     `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
