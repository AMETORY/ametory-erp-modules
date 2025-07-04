package models

import "github.com/AMETORY/ametory-erp-modules/shared"

type Citizen struct {
	shared.BaseModel
	FullName string `json:"full_name,omitempty"`
	NIK      string `gorm:"uniqueIndex;index" json:"nik,omitempty"`
	Address  string `json:"address,omitempty"`
	Phone    string `json:"phone,omitempty"`
}
