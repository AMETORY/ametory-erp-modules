package models

import (
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BranchModel struct {
	shared.BaseModel
	Name      string          `json:"name,omitempty"`
	Address   string          `json:"address,omitempty"`
	Employees []EmployeeModel `gorm:"foreignKey:BranchID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"employees,omitempty"`
	CompanyID string          `json:"company_id,omitempty" gorm:"not null"`
	Company   CompanyModel    `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE"`
}

func (b *BranchModel) TableName() string {
	return "branches"
}

func (b *BranchModel) BeforeCreate(tx *gorm.DB) (err error) {

	if b.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}
