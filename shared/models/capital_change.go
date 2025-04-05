package models

type CapitalChangeReport struct {
	GeneralReport
	OpeningBalance       float64 `json:"opening_balance"`
	ProfitLoss           float64 `json:"profit_loss"`
	PrivedBalance        float64 `json:"prived_balance"`
	CapitalChangeBalance float64 `json:"capital_change_balance"`
	EndingBalance        float64 `json:"ending_balance"`
}
