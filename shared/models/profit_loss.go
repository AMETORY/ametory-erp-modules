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
