package report_object

import "time"

type GeneralReport struct {
	Title        string    `json:"title" form:"title"`
	StartDate    time.Time `json:"start_date" form:"start_date"`
	EndDate      time.Time `json:"end_date" form:"end_date"`
	CurrencyCode string    `json:"currency_code" example:"currency_code"`
	CompanyID    string    `json:"company_id" example:"currency_code"`
}
