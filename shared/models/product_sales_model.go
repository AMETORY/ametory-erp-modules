package models

type ProductSalesCustomer struct {
	ProductID     string  `json:"product_id" gorm:"column:product_id"`
	ContactID     string  `json:"contact_id" gorm:"column:contact_id"`
	ProductCode   string  `json:"product_code" gorm:"column:product_code"`
	ContactCode   string  `json:"contact_code" gorm:"column:contact_code"`
	UnitCode      string  `json:"unit_code" gorm:"column:unit_code"`
	UnitName      string  `json:"unit_name" gorm:"column:unit_name"`
	ProductName   string  `json:"product_name" gorm:"column:product_name"`
	ContactName   string  `json:"contact_name" gorm:"column:contact_name"`
	Quantity      float64 `json:"quantity" gorm:"column:quantity"`
	TotalQuantity float64 `json:"total_quantity" gorm:"column:total_quantity"`
	TotalPrice    float64 `json:"total_price" gorm:"column:total_price"`
}
