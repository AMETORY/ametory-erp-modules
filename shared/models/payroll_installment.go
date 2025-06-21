package models

import (
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PayRollInstallment struct {
	shared.BaseModel
	CompanyID         *string       `json:"company_id" binding:"required"`
	Company           *CompanyModel `gorm:"foreignKey:CompanyID"`
	EmployeeLoanID    string        `json:"employee_loan_id"`
	EmployeeLoan      LoanModel     `gorm:"foreignKey:EmployeeLoanID"`
	InstallmentAmount float64       `json:"installment_amount"` // Jumlah cicilan yang dibayar di payroll ini
	PayRollItemID     string        `json:"pay_roll_item_id" gorm:"type:char(36)"`
	PayRoll           PayRollModel  `gorm:"-"`
}

func (pi *PayRollInstallment) TableName() string {
	return "payroll_installments"
}

func (pi *PayRollInstallment) BeforeCreate(tx *gorm.DB) (err error) {

	if pi.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return nil
}
