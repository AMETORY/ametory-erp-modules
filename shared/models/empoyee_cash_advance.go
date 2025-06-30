package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
)

type EmployeeCashAdvance struct {
	shared.BaseModel
	CashAdvanceNumber      string             `json:"cash_advance_number"`
	EmployeeID             *string            `json:"employee_id"`
	Employee               *EmployeeModel     `gorm:"foreignKey:EmployeeID"`
	CompanyID              *string            `json:"company_id"`
	Company                *CompanyModel      `gorm:"foreignKey:CompanyID"`
	Amount                 float64            `json:"amount"`
	AmountRequested        float64            `json:"amount_requested"`
	RemainingAmount        float64            `json:"remaining_amount"`
	RefundStatus           string             `json:"refund_status" gorm:"default:'PENDING'"` // 'PENDING', 'REFUNDED', 'PARTIALLY_REFUNDED'
	DateRequested          time.Time          `json:"date_requested"`
	DateApprovedOrRejected *time.Time         `json:"date_approved_or_rejected"`
	Status                 string             `json:"status" gorm:"default:'DRAFT'"` // 'DRAFT', 'APPROVED', 'REJECTED', 'FINISHED'
	Notes                  string             `json:"notes"`
	CashAccountID          *string            `json:"cash_account_id"`
	CashAccount            *AccountModel      `gorm:"foreignKey:CashAccountID"`
	ExpenseAccountID       *string            `json:"expense_account_id"`
	ExpenseAccount         *AccountModel      `gorm:"foreignKey:ExpenseAccountID"`
	IncomeAccountID        *string            `json:"income_account_id"`
	IncomeAccount          *AccountModel      `gorm:"foreignKey:IncomeAccountID"`
	ApproverID             *string            `json:"approver_id"`
	Approver               *EmployeeModel     `gorm:"foreignKey:ApproverID"`
	Remarks                string             `json:"remarks"`
	FileRefund             *FileModel         `json:"file_refund" gorm:"-"`
	CashAdvanceUsages      []CashAdvanceUsage `json:"usages" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Transactions           []TransactionModel `json:"transactions" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type CashAdvanceUsage struct {
	shared.BaseModel
	Date        time.Time  `json:"date"`
	Description string     `json:"description"`
	Amount      float64    `json:"amount"`
	File        *FileModel `json:"file" gorm:"-"`
}
