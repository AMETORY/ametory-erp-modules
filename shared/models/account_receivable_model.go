package models

import "time"

type AccountReceivableLedgerReport struct {
	TotalDebitBefore   float64                   `json:"total_debit_before"`
	TotalDebit         float64                   `json:"total_debit"`
	TotalDebitAfter    float64                   `json:"total_debit_after"`
	TotalCreditBefore  float64                   `json:"total_credit_before"`
	TotalCredit        float64                   `json:"total_credit"`
	TotalCreditAfter   float64                   `json:"total_credit_after"`
	TotalBalanceBefore float64                   `json:"total_balance_before"`
	TotalBalance       float64                   `json:"total_balance"`
	TotalBalanceAfter  float64                   `json:"total_balance_after"`
	GrandTotalDebit    float64                   `json:"grand_total_debit"`
	GrandTotalCredit   float64                   `json:"grand_total_credit"`
	GrandTotalBalance  float64                   `json:"grand_total_balance"`
	Ledgers            []AccountReceivableLedger `json:"ledgers"`
	Contact            ContactModel              `json:"contact"`
}

type AccountReceivableLedger struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
	Debit       float64   `json:"debit"`
	Credit      float64   `json:"credit"`
	Balance     float64   `json:"balance"`
	RefID       string    `json:"ref_id"`
	Ref         string    `json:"ref"`
	RefType     string    `json:"ref_type"`
}
