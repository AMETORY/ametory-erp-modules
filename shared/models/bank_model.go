package models

import (
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BankModel struct {
	shared.BaseModel
	Name string `json:"name,omitempty"`
	Code string `json:"code,omitempty"`
}

func (bm *BankModel) TableName() string {
	return "banks"
}

func (bm *BankModel) BeforeCreate(tx *gorm.DB) (err error) {

	if bm.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}
