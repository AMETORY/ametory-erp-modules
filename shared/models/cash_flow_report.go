package models

type CashFlowReport struct {
	GeneralReport
	Operating      []CashflowSubGroup `json:"operating"`
	Investing      []CashflowSubGroup `json:"investing"`
	Financing      []CashflowSubGroup `json:"financing"`
	TotalOperating float64            `json:"total_operating"`
	TotalInvesting float64            `json:"total_investing"`
	TotalFinancing float64            `json:"total_financing"`
}

type CashflowSubGroup struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
}
