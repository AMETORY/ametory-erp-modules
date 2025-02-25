package models

import (
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TeamModel adalah model database untuk team
type TeamModel struct {
	shared.BaseModel
	Name    string        `json:"name"`
	Members []MemberModel `json:"members" gorm:"foreignKey:TeamID"`
}

func (TeamModel) TableName() string {
	return "teams"
}

func (t *TeamModel) BeforeCreate(tx *gorm.DB) (err error) {
	t.ID = uuid.New().String()
	return nil
}
