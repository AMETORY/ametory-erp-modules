package models

type BalanceSheetAccount struct {
	ID   string  `json:"id"`
	Code string  `json:"code"`
	Name string  `json:"name"`
	Type string  `json:"type"`
	Sum  float64 `json:"sum"`
	Link string  `json:"link"`
}

type BalanceSheet struct {
	GeneralReport
	FixedAssets               []BalanceSheetAccount `json:"fixed_assets"`
	TotalFixed                float64               `json:"total_fixed"`
	CurrentAssets             []BalanceSheetAccount `json:"current_assets"`
	TotalCurrent              float64               `json:"total_current"`
	TotalAssets               float64               `json:"total_assets"`
	LiableAssets              []BalanceSheetAccount `json:"liable_assets"`
	TotalLiability            float64               `json:"total_liability"`
	Equity                    []BalanceSheetAccount `json:"equity"`
	TotalEquity               float64               `json:"total_equity"`
	TotalLiabilitiesAndEquity float64               `json:"total_liabilities_and_equity"`
}
