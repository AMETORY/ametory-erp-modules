package models

type BalanceAccount struct {
	ID             string           `json:"id"`
	Code           string           `json:"code"`
	Name           string           `json:"name"`
	Type           string           `json:"type"`
	Credit         float64          `json:"credit"`
	CreditCurrent  float64          `json:"credit_current"`
	CreditBefore   float64          `json:"credit_before"`
	Debit          float64          `json:"debit"`
	DebitCurrent   float64          `json:"debit_current"`
	DebitBefore    float64          `json:"debit_before"`
	Amount         float64          `json:"amount"`
	AmountCurrent  float64          `json:"amount_current"`
	AmountBefore   float64          `json:"amount_before"`
	Transactions   TransactionModel `json:"-"`
	IsSubAccount   bool             `json:"is_sub_account"`
	IsGroup        bool             `json:"is_group"`
	ShowZero       bool             `json:"show_zero"`
	Link           string           `json:"link"`
	IsSum          bool             `json:"is_sum"`
	IsTrialBalance bool             `json:"is_trial_balance"`
	CurrencyCode   string           `json:"currency_code" example:"currency_code"`
}
