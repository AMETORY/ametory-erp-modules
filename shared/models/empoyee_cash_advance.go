package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EmployeeCashAdvance struct {
	shared.BaseModel
	CashAdvanceNumber      string              `json:"cash_advance_number"`
	EmployeeID             *string             `json:"employee_id"`
	Employee               *EmployeeModel      `gorm:"foreignKey:EmployeeID"`
	CompanyID              *string             `json:"company_id"`
	Company                *CompanyModel       `gorm:"foreignKey:CompanyID"`
	Amount                 float64             `json:"amount"`
	AmountRequested        float64             `json:"amount_requested"`
	RemainingAmount        float64             `json:"remaining_amount"`
	RefundStatus           string              `json:"refund_status" gorm:"default:'PENDING'"` // 'PENDING', 'REFUNDED', 'PARTIALLY_REFUNDED'
	DateRequested          time.Time           `json:"date_requested"`
	DateApprovedOrRejected *time.Time          `json:"date_approved_or_rejected"`
	Status                 string              `json:"status" gorm:"default:'DRAFT'"` // 'DRAFT', 'APPROVED', 'REJECTED', 'FINISHED'
	Notes                  string              `json:"notes"`
	CashAccountID          *string             `json:"cash_account_id"`
	CashAccount            *AccountModel       `gorm:"foreignKey:CashAccountID"`
	ExpenseAccountID       *string             `json:"expense_account_id"`
	ExpenseAccount         *AccountModel       `gorm:"foreignKey:ExpenseAccountID"`
	IncomeAccountID        *string             `json:"income_account_id"`
	IncomeAccount          *AccountModel       `gorm:"foreignKey:IncomeAccountID"`
	ApproverID             *string             `json:"approver_id"`
	Approver               *EmployeeModel      `gorm:"foreignKey:ApproverID"`
	Remarks                string              `json:"remarks"`
	CashAdvanceUsages      []CashAdvanceUsage  `json:"usages" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:EmployeeCashAdvanceID"`
	Refunds                []CashAdvanceRefund `json:"refunds" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:EmployeeCashAdvanceID"`
	Transactions           []TransactionModel  `json:"transactions" gorm:"-"`
}

func (e *EmployeeCashAdvance) BeforeCreate(tx *gorm.DB) (err error) {

	if e.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return nil
}

type CashAdvanceUsage struct {
	shared.BaseModel
	EmployeeCashAdvanceID *string              `json:"employee_cash_advance_id"`
	EmployeeCashAdvance   *EmployeeCashAdvance `gorm:"foreignKey:EmployeeCashAdvanceID"`
	Date                  time.Time            `json:"date"`
	Description           string               `json:"description"`
	Amount                float64              `json:"amount"`
	Files                 []FileModel          `json:"files" gorm:"-"`
}

func (e *CashAdvanceUsage) BeforeCreate(tx *gorm.DB) (err error) {

	if e.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return nil
}

type CashAdvanceRefund struct {
	shared.BaseModel
	EmployeeCashAdvanceID *string              `json:"employee_cash_advance_id"`
	EmployeeCashAdvance   *EmployeeCashAdvance `gorm:"foreignKey:EmployeeCashAdvanceID"`
	Date                  time.Time            `json:"date"`
	Description           string               `json:"description"`
	Amount                float64              `json:"amount"`
	Files                 []FileModel          `json:"files" gorm:"-"`
}

func (e *CashAdvanceRefund) BeforeCreate(tx *gorm.DB) (err error) {

	if e.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return nil
}
