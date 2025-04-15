package models

type MonthlySalesReport struct {
	Year      int     `json:"year" sql:"year"`
	Month     int     `json:"month" sql:"month"`
	MonthName string  `json:"month_name" sql:"month_name"`
	WeekName  string  `json:"week_name" sql:"week_name"`
	Total     float64 `json:"total" sql:"total"`
	Company   string  `json:"company" sql:"company"`
}
