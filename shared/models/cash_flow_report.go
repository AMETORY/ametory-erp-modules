package models

import "github.com/AMETORY/ametory-erp-modules/shared/constants"

type CashFlowReport struct {
	GeneralReport
	Operating      []CashflowSubGroup `json:"operating,omitempty"`
	Investing      []CashflowSubGroup `json:"investing,omitempty"`
	Financing      []CashflowSubGroup `json:"financing,omitempty"`
	TotalOperating float64            `json:"total_operating,omitempty"`
	TotalInvesting float64            `json:"total_investing,omitempty"`
	TotalFinancing float64            `json:"total_financing,omitempty"`
}

type CashflowSubGroup struct {
	Name        string  `json:"name,omitempty"`
	Description string  `json:"description,omitempty"`
	Amount      float64 `json:"amount,omitempty"`
}

type CashflowGroupSetting struct {
	Operating []CashflowSubGroup `json:"operating,omitempty"`
	Investing []CashflowSubGroup `json:"investing,omitempty"`
	Financing []CashflowSubGroup `json:"financing,omitempty"`
}

func DefaultCasflowGroupSetting() CashflowGroupSetting {
	return CashflowGroupSetting{
		Operating: []CashflowSubGroup{
			{Name: constants.ACCEPTANCE_FROM_CUSTOMERS, Description: constants.ACCEPTANCE_FROM_CUSTOMERS_VALUE},
			{Name: constants.OTHER_CURRENT_ASSETS, Description: constants.OTHER_CURRENT_ASSETS_VALUE},
			{Name: constants.PAYMENT_TO_VENDORS, Description: constants.PAYMENT_TO_VENDORS_VALUE},
			{Name: constants.CREDIT_CARDS_AND_OTHER_SHORT_TERM_LIABILITIES, Description: constants.CREDIT_CARDS_AND_OTHER_SHORT_TERM_LIABILITIES_VALUE},
			{Name: constants.OTHER_INCOME, Description: constants.OTHER_INCOME_VALUE},
			{Name: constants.OPERATIONAL_EXPENSES, Description: constants.OPERATIONAL_EXPENSES_VALUE},
			{Name: constants.RETURNS_PAYMENT_OF_TAXES, Description: constants.RETURNS_PAYMENT_OF_TAXES_VALUE},
		},
		Investing: []CashflowSubGroup{
			{Name: constants.ACQUISITION_SALE_OF_ASSETS, Description: constants.ACQUISITION_SALE_OF_ASSETS_VALUE},
			{Name: constants.OTHER_INVESTMENT_ACTIVITIES, Description: constants.OTHER_INVESTMENT_ACTIVITIES_VALUE},
			{Name: constants.INVESTMENT_PARTNERSHIP, Description: constants.INVESTMENT_PARTNERSHIP_VALUE},
		},
		Financing: []CashflowSubGroup{
			{Name: constants.LOAN_PAYMENTS_RECEIPTS, Description: constants.LOAN_PAYMENTS_RECEIPTS_VALUE},
			{Name: constants.EQUITY_CAPITAL, Description: constants.EQUITY_CAPITAL_VALUE},
		},
	}
}

func CooperationCasflowGroupSetting() CashflowGroupSetting {
	setting := DefaultCasflowGroupSetting()
	setting.Operating = append(setting.Operating,
		CashflowSubGroup{Name: constants.COOPERATIVE_ACCEPTANCE_FROM_MEMBER, Description: constants.COOPERATIVE_ACCEPTANCE_FROM_MEMBER_LABEL},
		CashflowSubGroup{Name: constants.COOPERATIVE_ACCEPTANCE_FROM_NON_MEMBER, Description: constants.COOPERATIVE_ACCEPTANCE_FROM_NON_MEMBER_LABEL},
	)
	setting.Financing = append(setting.Financing,
		CashflowSubGroup{Name: constants.COOPERATIVE_PRINCIPAL_SAVING, Description: constants.COOPERATIVE_PRINCIPAL_SAVING_LABEL},
		CashflowSubGroup{Name: constants.COOPERATIVE_MANDATORY_SAVING, Description: constants.COOPERATIVE_MANDATORY_SAVING_LABEL},
		CashflowSubGroup{Name: constants.COOPERATIVE_VOLUNTARY_SAVING, Description: constants.COOPERATIVE_VOLUNTARY_SAVING_LABEL},
	)
	return setting
}

func ListSubGrup() []CashflowSubGroup {
	return []CashflowSubGroup{
		{Name: constants.ACCEPTANCE_FROM_CUSTOMERS, Description: constants.ACCEPTANCE_FROM_CUSTOMERS_VALUE},
		{Name: constants.OTHER_CURRENT_ASSETS, Description: constants.OTHER_CURRENT_ASSETS_VALUE},
		{Name: constants.PAYMENT_TO_VENDORS, Description: constants.PAYMENT_TO_VENDORS_VALUE},
		{Name: constants.CREDIT_CARDS_AND_OTHER_SHORT_TERM_LIABILITIES, Description: constants.CREDIT_CARDS_AND_OTHER_SHORT_TERM_LIABILITIES_VALUE},
		{Name: constants.OTHER_INCOME, Description: constants.OTHER_INCOME_VALUE},
		{Name: constants.OPERATIONAL_EXPENSES, Description: constants.OPERATIONAL_EXPENSES_VALUE},
		{Name: constants.RETURNS_PAYMENT_OF_TAXES, Description: constants.RETURNS_PAYMENT_OF_TAXES_VALUE},
		{Name: constants.COOPERATIVE_ACCEPTANCE_FROM_MEMBER, Description: constants.COOPERATIVE_ACCEPTANCE_FROM_MEMBER_LABEL},
		{Name: constants.COOPERATIVE_ACCEPTANCE_FROM_NON_MEMBER, Description: constants.COOPERATIVE_ACCEPTANCE_FROM_NON_MEMBER_LABEL},
		{Name: constants.ACQUISITION_SALE_OF_ASSETS, Description: constants.ACQUISITION_SALE_OF_ASSETS_VALUE},
		{Name: constants.OTHER_INVESTMENT_ACTIVITIES, Description: constants.OTHER_INVESTMENT_ACTIVITIES_VALUE},
		{Name: constants.INVESTMENT_PARTNERSHIP, Description: constants.INVESTMENT_PARTNERSHIP_VALUE},
		{Name: constants.LOAN_PAYMENTS_RECEIPTS, Description: constants.LOAN_PAYMENTS_RECEIPTS_VALUE},
		{Name: constants.EQUITY_CAPITAL, Description: constants.EQUITY_CAPITAL_VALUE},
		{Name: constants.COOPERATIVE_PRINCIPAL_SAVING, Description: constants.COOPERATIVE_PRINCIPAL_SAVING_LABEL},
		{Name: constants.COOPERATIVE_MANDATORY_SAVING, Description: constants.COOPERATIVE_MANDATORY_SAVING_LABEL},
		{Name: constants.COOPERATIVE_VOLUNTARY_SAVING, Description: constants.COOPERATIVE_VOLUNTARY_SAVING_LABEL},
	}
}
