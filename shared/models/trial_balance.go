package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
)

// TrialBalanceReport is a report for generate trial balance
type TrialBalanceReport struct {
	shared.BaseModel
	CompanyID    *string           `gorm:"type:char(36);index" json:"company_id"`
	Company      CompanyModel      `gorm:"foreignkey:CompanyID" json:"company,omitempty"`
	StartDate    time.Time         `json:"start_date"`
	EndDate      time.Time         `json:"end_date"`
	TrialBalance []TrialBalanceRow `json:"trial_balance,omitempty" gorm:"-"`
	Adjustment   []TrialBalanceRow `json:"adjustment,omitempty" gorm:"-"`
	BalanceSheet []TrialBalanceRow `json:"balance_sheet,omitempty" gorm:"-"`
	// TrialBalanceData string            `gorm:"type:JSON"`
	// AdjustmentData   string            `gorm:"type:JSON"`
	// BalanceSheetData string            `gorm:"type:JSON"`
}

// TrialBalanceRow is a row in trial balance report
type TrialBalanceRow struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	Debit   float64 `json:"debit"`
	Credit  float64 `json:"credit"`
	Balance float64 `json:"balance"`
	Code    string  `json:"code"`
}
