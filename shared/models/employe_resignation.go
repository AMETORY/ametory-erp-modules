package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EmployeeResignation struct {
	shared.BaseModel
	EmployeeID               *string        `json:"employee_id"`
	Employee                 *EmployeeModel `json:"employee" gorm:"foreignKey:EmployeeID"`
	UserID                   *string        `json:"user_id"`
	User                     *UserModel     `json:"user" gorm:"foreignKey:UserID"`
	CompanyID                *string        `json:"company_id" gorm:"not null"`
	Company                  *CompanyModel  `gorm:"foreignKey:CompanyID"`
	ResignationDate          time.Time      `gorm:"type:DATE;not null" json:"resignation_date"`
	ResignationEffectiveDate time.Time      `gorm:"type:DATE;not null" json:"resignation_effective_date"`
	Reason                   string         `gorm:"type:varchar(255);not null" json:"reason"`
	IsDeleted                bool           `gorm:"type:boolean;default:false" json:"-"`
	ApproverID               *string        `json:"approver_id"`
	Approver                 EmployeeModel  `json:"approver" gorm:"foreignKey:ApproverID"`
	ApprovalDate             *time.Time     `json:"approval_date"`
	ApprovalRemarks          *string        `json:"approval_remarks"`
	ApprovalStatus           string         `json:"approval_status" gorm:"default:'PENDING'"`
	Remarks                  string         `json:"remarks" gorm:"type:TEXT"`
	Files                    []FileModel    `json:"files" gorm:"-"`
}

func (e *EmployeeResignation) BeforeCreate(tx *gorm.DB) (err error) {
	if e.ID == "" {
		e.ID = uuid.New().String()
	}
	return nil
}
