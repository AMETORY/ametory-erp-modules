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
	Company                *CompanyModel        `gorm:"foreignKey:CompanyID"`
	EmployeeID             *string              `json:"employee_id"`
	Employee               *EmployeeModel       `gorm:"foreignKey:EmployeeID"`
	TotalAmount            float64              `json:"total_amount"`
	RemainingAmount        float64              `json:"remaining_amount"`
	InstallmentsPaid       float64              `json:"installments_paid"`
	Description            string               `json:"description"`
	Status                 string               `json:"status" gorm:"type:enum('APPROVED', 'ONGOING', 'PAID_OFF', 'CANCELLED');default:'APPROVED'"`
	PayRollInstallments    []PayRollInstallment `json:"pay_roll_installments" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"` // Relasi HasMany ke PayRollInstallment
	ApprovedID             *string              `json:"approved_id"`
	Approved               *UserModel           `gorm:"foreignKey:ApprovedID"`
	DateApprovedOrRejected *time.Time           `json:"date_approved_or_rejected"`
	Remarks                string               `json:"remarks"`
	File                   *string              `json:"file"`
	FileURL                *string              `json:"url" gorm:"-"`
}

func (u *EmployeeLoan) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}
