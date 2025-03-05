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
	Members []MemberModel `gorm:"many2many:team_members;"`
}

func (TeamModel) TableName() string {
	return "teams"
}

func (t *TeamModel) BeforeCreate(tx *gorm.DB) (err error) {

	if t.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return nil
}
