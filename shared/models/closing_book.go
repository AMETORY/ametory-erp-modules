package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
)

type ClosingBook struct {
	shared.BaseModel
	CompanyID         *string              `json:"company_id,omitempty"`
	Company           *CompanyModel        `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE" json:"company,omitempty"`
	UserID            *string              `json:"user_id,omitempty"`
	User              *UserModel           `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	StartDate         time.Time            `json:"start_date"`
	EndDate           time.Time            `json:"end_date"`
	Notes             string               `json:"notes"`
	Status            string               `json:"status"`
	ProfitLossData    *string              `gorm:"type:JSON" json:"profit_loss_data,omitempty"`
	ProfitLoss        *ProfitLossReport    `gorm:"-" json:"profit_loss"`
	BalanceSheetData  *string              `gorm:"type:JSON" json:"balance_sheet_data,omitempty"`
	BalanceSheet      *BalanceSheet        `gorm:"type:JSON" json:"balance_sheet"`
	CashFlowData      *string              `gorm:"type:JSON" json:"cash_flow_data,omitempty"`
	CashFlow          *CashFlowReport      `gorm:"type:JSON" json:"cash_flow"`
	TrialBalanceData  *string              `gorm:"type:JSON" json:"trial_balance_data,omitempty"`
	TrialBalance      *TrialBalanceReport  `gorm:"type:JSON" json:"trial_balance"`
	CapitalChangeData *string              `gorm:"type:JSON" json:"capital_change_data,omitempty"`
	CapitalChange     *CapitalChangeReport `gorm:"type:JSON" json:"capital_change"`
}
