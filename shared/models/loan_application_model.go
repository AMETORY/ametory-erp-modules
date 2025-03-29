package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LoanApplicationModel struct {
	shared.BaseModel
	CompanyID           *string                        `gorm:"size:36" json:"company_id"`
	UserID              *string                        `gorm:"size:36" json:"user_id"`
	MemberID            *string                        `json:"member_id"` // ID of the member submitting the application
	Member              *CooperativeMemberModel        `json:"member" gorm:"-"`
	LoanNumber          string                         `json:"loan_number"`
	LoanAmount          float64                        `json:"loan_amount"`                    // Amount of money requested
	LoanPurpose         string                         `json:"loan_purpose"`                   // Purpose of the loan (e.g., "modal usaha")
	LoanType            string                         `json:"loan_type"`                      // Type of loan ("Qardh Hasan", "Mudharabah", "Conventional")
	InterestRate        float64                        `json:"interest_rate,omitempty"`        // Interest rate for conventional loans, optional
	ExpectedProfitRate  float64                        `json:"expected_profit_rate,omitempty"` // Expected profit rate for Mudharabah, optional
	ProjectedProfit     float64                        `json:"projected_profit,omitempty"`     // Expected profit rate for Mudharabah, optional
	SubmissionDate      time.Time                      `json:"submission_date"`                // Date of loan application submission
	RepaymentTerm       int                            `json:"repayment_term"`                 // Loan repayment term in months
	Status              string                         `json:"status"`                         // Status of the application ("Pending", "Approved", "Rejected")
	ApprovedBy          *string                        `json:"approved_by,omitempty"`          // Name of the approver (optional)
	DisbursementDate    *time.Time                     `json:"disbursement_date,omitempty"`    // Date when the loan is disbursed (optional)
	Remarks             string                         `json:"remarks,omitempty"`              // Additional remarks or notes (optional)
	ProfitType          string                         `json:"profit_type"`                    // "fixed", "declining", or "effective" - type of profit/bunga
	AdminFee            float64                        `json:"admin_fee"`                      // Biaya administrasi
	AccountReceivableID *string                        `gorm:"size:36" json:"account_receivable_id"`
	AccountIncomeID     *string                        `gorm:"size:36" json:"account_income_id"`
	AccountAdminFeeID   *string                        `gorm:"size:36" json:"account_admin_fee_id"`
	AccountAssetID      *string                        `gorm:"size:36" json:"account_asset_id"`
	Data                string                         `json:"data" gorm:"type:JSON"`
	Preview             map[string][]InstallmentDetail `json:"preview" gorm:"-"`
	TermCondition       string                         `json:"term_condition" gorm:"type:TEXT"` // Terms and Conditions of the loan
	Payments            []InstallmentPayment           `json:"payments,omitempty" gorm:"-"`
	LastPayment         *InstallmentPayment            `json:"last_payment,omitempty" gorm:"-"`
	Transactions        []TransactionModel             `json:"transactions,omitempty" gorm:"-"`
	Installments        []InstallmentDetail            `json:"installments,omitempty" gorm:"-"`
	NetSurplusID        *string                        `gorm:"size:36" json:"net_surplus_id"`
}

func (LoanApplicationModel) TableName() string {
	return "loan_applications"
}

func (m *LoanApplicationModel) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	return
}
