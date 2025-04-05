package models

type ProfitLoss struct {
	GeneralReport
	Profit       ProfitLossCategory `json:"profit"`
	Loss         ProfitLossCategory `json:"loss"`
	NetSurplus   ProfitLossCategory `json:"net_surplus"`
	Amount       float64            `json:"amount"`
	IsPeriodical bool
	CurrencyCode string `json:"currency_code" example:"currency_code"`
}

type ProfitLossCategory struct {
	Title             string           `json:"title"`
	ProfitLossAccount []BalanceAccount `json:"accounts"`
	SubTotal          float64          `json:"subtotal"`
	CurrencyCode      string           `json:"currency_code" example:"currency_code"`
}

type ProfitLossReport struct {
	GeneralReport
	Profit       []ProfilLossAccount `json:"profit"`
	Loss         []ProfilLossAccount `json:"loss"`
	GrossProfit  float64             `json:"gross_profit"`
	TotalExpense float64             `json:"total_expense"`
	NetProfit    float64             `json:"net_profit"`
}

type ProfilLossAccount struct {
	ID   string  `json:"id"`
	Code string  `json:"code"`
	Name string  `json:"name"`
	Type string  `json:"type"`
	Sum  float64 `json:"sum"`
}
