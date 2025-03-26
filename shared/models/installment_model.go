package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
)

type InstallmentPayment struct {
	shared.BaseModel
	LoanApplicationID *string              `gorm:"size:30" json:"loan_application_id"`                                               // ID of the loan this payment belongs to
	LoanApplication   LoanApplicationModel `gorm:"foreignKey:LoanApplicationID;constraint:OnDelete:CASCADE" json:"loan_application"` // Loan application associated with the payment
	MemberID          *string              `json:"member_id"`                                                                        // ID of the member making the payment
	InstallmentNo     int                  `json:"installment_no"`                                                                   // Installment number (e.g., 1, 2, 3, ...)
	PaymentDate       time.Time            `json:"payment_date"`                                                                     // Date of payment
	PrincipalPaid     float64              `json:"principal_paid"`                                                                   // Amount paid towards the principal
	ProfitPaid        float64              `json:"profit_paid"`                                                                      // Amount paid as profit (interest or share)
	AdminFeePaid      float64              `json:"admin_fee_paid"`                                                                   // Administrative fee paid (if any)
	TotalPaid         float64              `json:"total_paid"`                                                                       // Total payment made (sum of principal, profit, admin fee)
	RemainingLoan     float64              `json:"remaining_loan"`                                                                   // Remaining loan balance after the payment
	PaymentAmount     float64              `json:"payment_amount"`
	Remarks           string               `json:"remarks,omitempty"` // Additional remarks or notes (optional)
}

type InstallmentDetail struct {
	InstallmentNumber int     `json:"installment_number"` // Nomor cicilan
	PrincipalAmount   float64 `json:"principal_amount"`   // Angsuran pokok
	InterestAmount    float64 `json:"interest_amount"`    // Bunga per cicilan
	AdminFee          float64 `json:"admin_fee"`          // Biaya administrasi
	TotalPaid         float64 `json:"total_paid"`         // Total pembayaran (pokok + bunga + admin)
	RemainingLoan     float64 `json:"remaining_loan"`     // Sisa pinjaman setelah cicilan
}
