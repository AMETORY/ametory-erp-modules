package transaction

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/company"
	"github.com/AMETORY/ametory-erp-modules/finance/account"
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TransactionModel struct {
	shared.BaseModel
	Code                   string                `json:"code"`
	Description            string                `json:"description"`
	Notes                  string                `json:"notes"`
	Credit                 float64               `json:"credit"`
	Debit                  float64               `json:"debit"`
	Amount                 float64               `json:"amount"`
	Date                   time.Time             `json:"date"`
	IsOpeningBalance       bool                  `json:"is_opening_balance"`
	IsIncome               bool                  `json:"is_income"`
	IsExpense              bool                  `json:"is_expense"`
	IsJournal              bool                  `json:"is_journal"`
	IsRefund               bool                  `json:"is_refund"`
	IsAccountReceivable    bool                  `json:"is_account_receivable"`
	IsAccountPayable       bool                  `json:"is_account_payable"`
	AccountID              *string               `json:"account_id"`
	Account                account.AccountModel  `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;foreignKey:AccountID" json:"account"`
	TaxPaymentID           string                `json:"tax_payment_id,omitempty"`
	TransactionRefID       *string               `json:"transaction_ref_id,omitempty"`
	TransactionRefType     string                `json:"transaction_ref_type,omitempty"`
	TransactionRefs        []TransactionModel    `json:"transaction_refs,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;foreignKey:TransactionRefID"`
	CompanyID              *string               `json:"company_id,omitempty" gorm:"null"`
	Company                *company.CompanyModel `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	File                   *string               `json:"file,omitempty"`
	FileURL                *string               `json:"url,omitempty" gorm:"-"`
	SourceID               *string               `json:"source_id,omitempty" gorm:"-"`
	DestinationID          *string               `json:"destination_id,omitempty" gorm:"-"`
	IsTakeHomePay          bool                  `json:"is_take_home_pay,omitempty"`
	PayRollPayableID       string                `json:"pay_roll_payable_id,omitempty"`
	IsPayRollPayment       bool                  `json:"is_pay_roll_payment,omitempty"`
	IsReimbursementPayment bool                  `json:"is_reimbursement_payment,omitempty"`
	IsEmployeeLoanPayment  bool                  `json:"is_employee_loan_payment,omitempty"`
	// EmployeeID             *string              `json:"employee_id"`
	// Employee               Employee             `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;foreignKey:EmployeeID" json:"-"`
	// Images                 []Image            `json:"images" gorm:"-"`
	// PayRollID              *string            `json:"pay_roll_id"`
	// PayRoll                PayRoll            `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;foreignKey:PayRollID" json:"-"`
	// ReimbursementID        *string            `json:"reimbursement_id"`
	// Reimbursement          Reimbursement      `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;foreignKey:ReimbursementID" json:"-"`
	// CashAdvanceID          *string            `json:"cash_advance_id"`
	// CashAdvance            CashAdvance        `gorm:"foreignKey:CashAdvanceID"`
	// EmployeeLoanID         *string            `json:"employee_loan_id"`
	// EmployeeLoan           EmployeeLoan       `gorm:"foreignKey:EmployeeLoanID"`
	// PayRollInstallmentID   *string            `json:"pay_roll_installment_id"`
	// PayRollInstallment     PayRollInstallment `gorm:"foreignKey:PayRollInstallmentID"`

}

func (TransactionModel) TableName() string {
	return "transactions"
}
func (u *TransactionModel) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&TransactionModel{})
}
