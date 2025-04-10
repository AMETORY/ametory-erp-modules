package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
)

type ClosingBook struct {
	shared.BaseModel
	CompanyID        *string             `gorm:"type:char(36);index" json:"company_id"`
	Company          CompanyModel        `gorm:"foreignkey:CompanyID" json:"company,omitempty"`
	StartDate        time.Time           `json:"start_date"`
	EndDate          time.Time           `json:"end_date"`
	Notes            string              `json:"notes"`
	Status           string              `json:"status"`
	ProfitLossData   *string             `gorm:"type:JSON" json:"profit_loss_data,omitempty"`
	ProfitLoss       *ProfitLossReport   `gorm:"-" json:"profit_loss"`
	BalanceSheetData *string             `gorm:"type:JSON" json:"balance_sheet_data,omitempty"`
	BalanceSheet     *BalanceSheet       `gorm:"type:JSON" json:"balance_sheet"`
	CashFlowData     *string             `gorm:"type:JSON" json:"cash_flow_data,omitempty"`
	CashFlow         *CashFlowReport     `gorm:"type:JSON" json:"cash_flow"`
	TrialBalanceData *string             `gorm:"type:JSON" json:"trial_balance_data,omitempty"`
	TrialBalance     *TrialBalanceReport `gorm:"type:JSON" json:"trial_balance"`
}
