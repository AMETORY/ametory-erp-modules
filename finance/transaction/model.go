package transaction

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/company"
	"github.com/AMETORY/ametory-erp-modules/finance/account"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TransactionModel struct {
	utils.BaseModel
	Description            string               `json:"description"`
	Notes                  string               `json:"notes"`
	Credit                 float64              `json:"credit"`
	Debit                  float64              `json:"debit"`
	Amount                 float64              `json:"amount"`
	Date                   time.Time            `json:"date"`
	IsOpeningBalance       bool                 `json:"is_opening_balance"`
	IsIncome               bool                 `json:"is_income"`
	IsExpense              bool                 `json:"is_expense"`
	IsJournal              bool                 `json:"is_journal"`
	IsRefund               bool                 `json:"is_refund"`
	IsAccountReceivable    bool                 `json:"is_account_receivable"`
	IsAccountPayable       bool                 `json:"is_account_payable"`
	AccountSourceID        *string              `json:"account_source_id"`
	AccountSourceName      string               `json:"account_source_name" gorm:"-"`
	AccountSource          account.AccountModel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;foreignKey:AccountSourceID" json:"-"`
	AccountDestinationID   *string              `json:"account_destination_id"`
	AccountDestinationName string               `json:"account_destination_name" gorm:"-"`
	AccountDestination     account.AccountModel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;foreignKey:AccountDestinationID" json:"-"`
	TaxPaymentID           string               `json:"tax_payment_id"`
	PayRollPayableID       string               `json:"pay_roll_payable_id"`
	IsPayRollPayment       bool                 `json:"is_pay_roll_payment"`
	IsReimbursementPayment bool                 `json:"is_reimbursement_payment"`
	IsEmployeeLoanPayment  bool                 `json:"is_employee_loan_payment"`
	TransactionRefID       *string              `json:"transaction_ref_id"`
	TransactionRefs        []TransactionModel   `json:"transaction_refs" gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;foreignKey:TransactionRefID"`
	IsTakeHomePay          bool                 `json:"is_take_home_pay"`
	CompanyID              string               `json:"company_id" gorm:"not null"`
	Company                company.CompanyModel `gorm:"foreignKey:CompanyID"`
	File                   *string              `json:"file"`
	FileURL                *string              `json:"url" gorm:"-"`
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

func (u *TransactionModel) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&TransactionModel{})
}
