package models

import (
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PayrollItemModel struct {
	shared.BaseModel
	ItemType           string              `gorm:"type:varchar(20);not null" json:"item_type" ` //'SALARY', 'ALLOWANCE', 'OVERTIME', 'DEDUCTION', 'REIMBURSEMENT'
	AccountPayableID   *string             `json:"account_payable_id"`
	Title              string              `json:"title"`
	Notes              string              `json:"notes"`
	IsDefault          bool                `json:"is_default"`
	IsDeductible       bool                `json:"is_deductible"`
	IsTax              bool                `json:"is_tax"`
	TaxAutoCount       bool                `json:"tax_auto_count"`
	IsTaxCost          bool                `json:"is_tax_cost"`
	IsTaxAllowance     bool                `json:"is_tax_allowance"`
	Amount             float64             `json:"amount"`
	PayRollID          string              `json:"pay_roll_id"`
	PayRoll            PayRollModel        `gorm:"foreignKey:PayRollID;constraint:OnDelete:CASCADE" json:"-"`
	ReimbursementID    *string             `json:"reimbursement_id"`
	Reimbursement      ReimbursementModel  `gorm:"foreignKey:ReimbursementID" json:"-"`
	Bpjs               bool                `json:"bpjs"`
	BpjsCounted        bool                `json:"bpjs_counted"`
	Tariff             float64             `json:"tariff"`
	CompanyID          string              `json:"company_id" gorm:"not null"`
	Company            CompanyModel        `gorm:"foreignKey:CompanyID"`
	Data               string              `gorm:"type:JSON" json:"data"`
	EmployeeLoanID     *string             `json:"employee_loan_id,omitempty"`
	EmployeeLoan       *LoanModel          `gorm:"foreignKey:EmployeeLoanID;constraint:OnDelete:CASCADE"`
	PayRollInstallment *PayRollInstallment `gorm:"foreignKey:PayRollItemID"`
}

func (PayrollItemModel) TableName() string {
	return "payroll_items"
}

func (pi *PayrollItemModel) BeforeCreate(tx *gorm.DB) error {

	if pi.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return nil
}
