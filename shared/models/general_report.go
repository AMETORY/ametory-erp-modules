package models

import "time"

type GeneralReport struct {
	Title        string    `json:"title,omitempty" form:"title"`
	StartDate    time.Time `json:"start_date,omitempty" form:"start_date"`
	EndDate      time.Time `json:"end_date,omitempty" form:"end_date"`
	CurrencyCode string    `json:"currency_code,omitempty" example:"currency_code"`
	CompanyID    string    `json:"company_id,omitempty" example:"currency_code"`
}
