package models

import (
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrganizationModel struct {
	shared.BaseModel
	Parent           *OrganizationModel
	ParentId         *string             `gorm:"REFERENCES organizations" json:"parent_id,omitempty"`
	Name             string              `gorm:"size:36" json:"name,omitempty"`
	Code             string              `gorm:"size:36" json:"code,omitempty"`
	Description      string              `gorm:"size:100" json:"description,omitempty"`
	Employees        []EmployeeModel     `gorm:"foreignKey:organization_id;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	SubOrganizations []OrganizationModel `json:"sub_organizations,omitempty" gorm:"foreignKey:parent_id"`
	CompanyID        *string             `json:"company_id" gorm:"not null"`
	Company          *CompanyModel       `gorm:"foreignKey:CompanyID"`
}

func (o OrganizationModel) TableName() string {
	return "organizations"
}

func (o *OrganizationModel) BeforeCreate(tx *gorm.DB) (err error) {

	if o.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}
