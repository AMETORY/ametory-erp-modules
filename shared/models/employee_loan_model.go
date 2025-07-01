package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EmployeeLoan struct {
	shared.BaseModel
	LoanNumber             string               `json:"loan_number"`
	CompanyID              *string              `json:"company_id"`
	Company                *CompanyModel        `gorm:"foreignKey:CompanyID" json:"company"`
	EmployeeID             *string              `json:"employee_id"`
	Employee               *EmployeeModel       `gorm:"foreignKey:EmployeeID" json:"employee"`
	TotalAmount            float64              `json:"total_amount"`
	Date                   time.Time            `json:"date"`
	RemainingAmount        float64              `json:"remaining_amount"`
	InstallmentsPaid       float64              `json:"installments_paid"`
	Description            string               `json:"description"`
	Status                 string               `json:"status" gorm:"default:'DRAFT'"`
	PayRollInstallments    []PayRollInstallment `json:"pay_roll_installments" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"` // Relasi HasMany ke PayRollInstallment
	ApproverID             *string              `json:"approver_id"`
	Approver               *EmployeeModel       `gorm:"foreignKey:ApproverID" json:"approver"`
	DateApprovedOrRejected *time.Time           `json:"date_approved_or_rejected"`
	Remarks                string               `json:"remarks"`
	File                   *FileModel           `json:"file" gorm:"-"`
	ApprovalByAdminID      *string              `json:"approval_by_admin_id"`
	ApprovalByAdmin        *UserModel           `json:"approval_by_admin" gorm:"foreignKey:ApprovalByAdminID"`
}

func (u *EmployeeLoan) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}
