package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LoanModel struct {
	shared.BaseModel
	LoanNumber             string               `json:"loan_number"`
	CompanyID              string               `json:"company_id"`
	Company                CompanyModel         `gorm:"foreignKey:CompanyID"`
	EmployeeID             *string              `json:"employee_id"`
	Employee               *EmployeeModel       `gorm:"foreignKey:EmployeeID"`
	TotalAmount            float64              `json:"total_amount"`
	RemainingAmount        float64              `json:"remaining_amount"`
	InstallmentsPaid       float64              `json:"installments_paid"`
	Description            string               `json:"description"`
	Status                 string               `json:"status" gorm:"default:'PENDING'"`
	PayRollInstallments    []PayRollInstallment `json:"pay_roll_installments" gorm:"foreignKey:EmployeeLoanID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"` // Relasi HasMany ke PayRollInstallment
	ApprovedID             *string              `json:"approved_id"`
	Approved               *UserModel           `gorm:"foreignKey:ApprovedID;constraint:OnDelete:CASCADE;"`
	DateApprovedOrRejected *time.Time           `json:"date_approved_or_rejected"`
	Remarks                string               `json:"remarks"`
	FileID                 *string              `json:"file_id" gorm:"type:char(36)"`
	File                   *FileModel           `json:"file"`
}

func (LoanModel) TableName() string {
	return "loans"
}

func (m *LoanModel) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	return
}
