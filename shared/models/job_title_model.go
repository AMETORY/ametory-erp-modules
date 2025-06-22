package models

import (
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type JobTitleModel struct {
	shared.BaseModel
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	// Employees []EmployeeModel `json:"employees,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Employees []EmployeeModel `json:"employees,omitempty" gorm:"foreignKey:JobTitleID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	CompanyID *string         `json:"company_id,omitempty" gorm:"not null"`
	Company   *CompanyModel   `json:"company,omitempty" gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE"`
}

func (j *JobTitleModel) TableName() string {
	return "job_titles"
}

func (j *JobTitleModel) BeforeCreate(tx *gorm.DB) (err error) {

	if j.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}
