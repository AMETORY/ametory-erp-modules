package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
)

// AnnoucementModel adalah model database untuk menampung data announcement
type AnnoucementModel struct {
	shared.BaseModel
	EffectiveDate  *time.Time              `gorm:"index" json:"effective_date"`
	EffectiveUntil *time.Time              `gorm:"index" json:"effective_until"`
	CompanyID      *string                 `gorm:"index" json:"company_id"`
	Company        *CompanyModel           `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE" json:"company,omitempty"`
	BranchID       *string                 `gorm:"index" json:"branch_id"`
	Branch         *BranchModel            `gorm:"foreignKey:BranchID;constraint:OnDelete:CASCADE" json:"branch,omitempty"`
	OrganizationID *string                 `gorm:"index" json:"organization_id"`
	Organization   *OrganizationModel      `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE" json:"organization,omitempty"`
	JobTitleID     *string                 `gorm:"index" json:"job_title_id"`
	JobTitle       *JobTitleModel          `gorm:"foreignKey:JobTitleID;constraint:OnDelete:CASCADE" json:"job_title,omitempty"`
	Employees      []EmployeeActivityModel `gorm:"many2many:announcement_employees;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"employees,omitempty"`
	Title          string                  `gorm:"type:varchar(255)" json:"title"`
	Content        string                  `gorm:"type:text" json:"content"`
	FileModels     []FileModel             `gorm:"-" json:"files,omitempty"`
}
