package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EmployeeCashAdvance struct {
	shared.BaseModel
	CashAdvanceNumber            string              `json:"cash_advance_number,omitempty"`
	EmployeeID                   *string             `json:"employee_id,omitempty"`
	Employee                     *EmployeeModel      `gorm:"foreignKey:EmployeeID" json:"employee"`
	CompanyID                    *string             `json:"company_id,omitempty"`
	Company                      *CompanyModel       `gorm:"foreignKey:CompanyID" json:"company"`
	Amount                       float64             `json:"amount,omitempty"`
	Balance                      float64             `json:"balance" default:"amount"` // Balance after cash advance
	AmountRequested              float64             `json:"amount_requested,omitempty"`
	RemainingAmount              float64             `json:"remaining_amount,omitempty"`
	RefundStatus                 string              `json:"refund_status,omitempty" gorm:"default:'PENDING'"` // 'PENDING', 'REFUNDED', 'PARTIALLY_REFUNDED'
	DateRequested                time.Time           `json:"date_requested,omitempty"`
	DateApprovedOrRejected       *time.Time          `json:"date_approved_or_rejected,omitempty"`
	Status                       string              `json:"status,omitempty" gorm:"default:'DRAFT'"` // 'DRAFT', 'APPROVED', 'REJECTED', 'FINISHED'
	Notes                        string              `json:"notes,omitempty"`
	CashAccountID                *string             `json:"cash_account_id,omitempty"`
	CashAccount                  *AccountModel       `gorm:"foreignKey:CashAccountID"`
	ExpenseAccountID             *string             `json:"expense_account_id,omitempty"`
	ExpenseAccount               *AccountModel       `gorm:"foreignKey:ExpenseAccountID"`
	IncomeAccountID              *string             `json:"income_account_id,omitempty"`
	IncomeAccount                *AccountModel       `gorm:"foreignKey:IncomeAccountID"`
	ApproverID                   *string             `json:"approver_id,omitempty"`
	Approver                     *EmployeeModel      `gorm:"foreignKey:ApproverID" json:"approver"`
	Remarks                      string              `json:"remarks,omitempty"`
	CashAdvanceUsages            []CashAdvanceUsage  `json:"usages,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:EmployeeCashAdvanceID"`
	Refunds                      []CashAdvanceRefund `json:"refunds,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:EmployeeCashAdvanceID"`
	Transactions                 []TransactionModel  `json:"transactions,omitempty" gorm:"-"`
	ApprovalByAdminID            *string             `json:"approval_by_admin_id"`
	ApprovalByAdmin              *UserModel          `json:"approval_by_admin" gorm:"foreignKey:ApprovalByAdminID"`
	RefundRemarks                string              `json:"refund_remarks,omitempty"`
	RefundDateApprovedOrRejected *time.Time          `json:"refund_date_approved_or_rejected,omitempty"`
	RefundApprovalByAdminID      *string             `json:"refund_approval_by_admin_id"`
	RefundApprovalByAdmin        *UserModel          `json:"refund_approval_by_admin" gorm:"foreignKey:RefundApprovalByAdminID"`
	File                         *FileModel          `json:"file" gorm:"-"`
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
