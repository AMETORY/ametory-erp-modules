package report_object

type Category struct {
	Title           string           `json:"title"`
	BalanceAccount  []BalanceAccount `json:"balance_accounts"`
	SubTotal        float64          `json:"subtotal"`
	SubTotalCurrent float64          `json:"subtotal_current"`
	SubTotalBefore  float64          `json:"subtotal_before"`
	Credit          float64          `json:"credit"`
	Debit           float64          `json:"debit"`
	CreditBefore    float64          `json:"credit_before"`
	DebitBefore     float64          `json:"debit_before"`
	CreditCurrent   float64          `json:"credit_current"`
	DebitCurrent    float64          `json:"debit_current"`
	CurrencyCode    string           `json:"currency_code" example:"currency_code"`
}
