package models

import (
	"encoding/json"
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"gorm.io/gorm"
)

type ClosingBook struct {
	shared.BaseModel
	CompanyID          *string              `json:"company_id,omitempty"`
	Company            *CompanyModel        `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE" json:"company,omitempty"`
	UserID             *string              `json:"user_id,omitempty"`
	User               *UserModel           `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	StartDate          time.Time            `json:"start_date"`
	EndDate            time.Time            `json:"end_date"`
	Notes              string               `json:"notes"`
	Status             string               `json:"status"`
	ProfitLossData     *string              `gorm:"type:JSON" json:"profit_loss_data,omitempty"`
	ProfitLoss         *ProfitLossReport    `gorm:"-" json:"profit_loss"`
	BalanceSheetData   *string              `gorm:"type:JSON" json:"balance_sheet_data,omitempty"`
	BalanceSheet       *BalanceSheet        `gorm:"-" json:"balance_sheet"`
	CashFlowData       *string              `gorm:"type:JSON" json:"cash_flow_data,omitempty"`
	CashFlow           *CashFlowReport      `gorm:"-" json:"cash_flow"`
	TrialBalanceData   *string              `gorm:"type:JSON" json:"trial_balance_data,omitempty"`
	TrialBalance       *TrialBalanceReport  `gorm:"-" json:"trial_balance"`
	CapitalChangeData  *string              `gorm:"type:JSON" json:"capital_change_data,omitempty"`
	CapitalChange      *CapitalChangeReport `gorm:"-" json:"capital_change"`
	Transactions       []TransactionModel   `json:"transactions,omitempty" gorm:"-"`
	TransactionData    *string              `gorm:"type:JSON" json:"transaction_data,omitempty"`
	ClosingSummaryData *string              `gorm:"type:JSON" json:"closing_summary_data,omitempty"`
	ClosingSummary     *ClosingSummary      `gorm:"-" json:"closing_summary"`
}

func (c *ClosingBook) AfterFind(tx *gorm.DB) (err error) {

	if c.BalanceSheetData != nil {
		if err := json.Unmarshal([]byte(*c.BalanceSheetData), &c.BalanceSheet); err != nil {
			return err
		}
	}
	if c.CashFlowData != nil {
		if err := json.Unmarshal([]byte(*c.CashFlowData), &c.CashFlow); err != nil {
			return err
		}
	}
	if c.TrialBalanceData != nil {
		if err := json.Unmarshal([]byte(*c.TrialBalanceData), &c.TrialBalance); err != nil {
			return err
		}
	}
	if c.CapitalChangeData != nil {
		if err := json.Unmarshal([]byte(*c.CapitalChangeData), &c.CapitalChange); err != nil {
			return err
		}
	}
	if c.TransactionData != nil {
		if err := json.Unmarshal([]byte(*c.TransactionData), &c.Transactions); err != nil {
			return err
		}
	}
	if c.ClosingSummaryData != nil {
		if err := json.Unmarshal([]byte(*c.ClosingSummaryData), &c.ClosingSummary); err != nil {
			return err
		}
		if c.ClosingSummary.TaxExpenseID != nil {
			if err := tx.Where("id = ?", c.ClosingSummary.TaxExpenseID).First(&c.ClosingSummary.TaxExpense).Error; err != nil {
				return err
			}
		}
		if c.ClosingSummary.TaxPayableID != nil {
			if err := tx.Where("id = ?", c.ClosingSummary.TaxPayableID).First(&c.ClosingSummary.TaxPayable).Error; err != nil {
				return err
			}
		}
		if c.ClosingSummary.EarningRetainID != nil {
			if err := tx.Where("id = ?", c.ClosingSummary.EarningRetainID).First(&c.ClosingSummary.EarningRetain).Error; err != nil {
				return err
			}
		}
	}
	if c.ProfitLossData != nil {
		if err := json.Unmarshal([]byte(*c.ProfitLossData), &c.ProfitLoss); err != nil {
			return err
		}

		if c.ClosingSummary.IncomeTax > 0 {
			c.ProfitLoss.Tax = append(c.ProfitLoss.Tax, ProfitLossAccount{
				Name: "Pajak",
				Sum:  c.ClosingSummary.IncomeTax,
			})
			c.ProfitLoss.IncomeTax = c.ClosingSummary.IncomeTax

			c.ProfitLoss.NetProfitAfterTax = c.ProfitLoss.NetProfit - c.ClosingSummary.IncomeTax
		}
	}
	return nil
}

type ClosingSummary struct {
	TotalIncome     float64       `json:"total_income,omitempty"`
	TotalExpense    float64       `json:"total_expense,omitempty"`
	NetIncome       float64       `json:"net_income,omitempty"`
	IncomeTax       float64       `json:"income_tax,omitempty"`
	TaxPayableID    *string       `json:"tax_payable_id,omitempty"`
	TaxExpenseID    *string       `json:"tax_expense_id,omitempty"`
	TaxPercentage   float64       `json:"tax_percentage,omitempty"`
	TaxPayable      *AccountModel `json:"tax_payable,omitempty"`
	TaxExpense      *AccountModel `json:"tax_expense,omitempty"`
	EarningRetainID *string       `json:"earning_retain_id,omitempty"`
	EarningRetain   *AccountModel `json:"earning_retain,omitempty"`
}
