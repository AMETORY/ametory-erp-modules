package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReimbursementModel struct {
	shared.BaseModel
	AccountPayableID *string                  `gorm:"size:36" json:"account_payable_id"`
	AccountPayable   *AccountModel            `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;foreignKey:AccountPayableID" json:"-"`
	AccountExpenseID *string                  `gorm:"size:36" json:"account_expense_id"`
	AccountExpense   AccountModel             `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;foreignKey:AccountExpenseID" json:"-"`
	Date             time.Time                `json:"date" binding:"required"`
	Name             string                   `json:"name"`
	Notes            string                   `json:"notes"`
	Remarks          string                   `json:"remarks"`
	Total            float64                  `json:"total"`
	Balance          float64                  `json:"balance"`
	Status           string                   `json:"status" gorm:"default:'DRAFT'"`
	Items            []ReimbursementItemModel `json:"items" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	EmployeeID       string                   `json:"employee_id"`
	Employee         EmployeeModel            `gorm:"foreignKey:EmployeeID" json:"-"`
	Transactions     []TransactionModel       `json:"transactions" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Attachment       string                   `json:"attachment"`
	CompanyID        string                   `json:"company_id" gorm:"not null"`
	Company          CompanyModel             `gorm:"foreignKey:CompanyID"`
}

func (ReimbursementModel) TableName() string {
	return "reimbursements"
}

func (r *ReimbursementModel) BeforeCreate(tx *gorm.DB) (err error) {

	if r.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}
