package models

import (
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PayRollCostModel struct {
	shared.BaseModel
	Description   string
	PayRollID     string           `json:"pay_roll_id"`
	PayRoll       PayRollModel     `gorm:"foreignKey:PayRollID" json:"-"`
	PayRollItemID string           `json:"pay_roll_item_id"`
	PayRollItem   PayrollItemModel `gorm:"foreignKey:PayRollItemID" json:"-"`
	Amount        float64          `json:"amount"`
	Tariff        float64          `json:"tariff"`
	BpjsTkJht     bool             `json:"bpjs_tk_jht"`
	BpjsTkJp      bool             `json:"bpjs_tk_jp"`
	DebtDeposit   bool             `json:"debt_deposit"`
	CompanyID     string           `json:"company_id" gorm:"not null"`
	Company       CompanyModel     `gorm:"foreignKey:CompanyID"`
}

func (PayRollCostModel) TableName() string {
	return "payroll_costs"
}

func (pc *PayRollCostModel) BeforeCreate(tx *gorm.DB) error {

	if pc.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return nil
}
