package models

import "github.com/AMETORY/ametory-erp-modules/shared"

type Feature struct {
	shared.BaseModel
	Name        string `json:"name"`
	Description string `json:"description"`
	OrderNumber int    `json:"order_number"`
	Code        string `json:"code" gorm:"uniqueIndex"`
}
