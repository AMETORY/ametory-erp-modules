package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NetSurplusModel struct {
	shared.BaseModel
	CompanyID        *string                `gorm:"size:36" json:"company_id"`
	UserID           *string                `gorm:"size:36" json:"user_id"`
	StartDate        time.Time              `json:"start_date"`
	EndDate          time.Time              `json:"end_date"`
	Date             time.Time              `json:"date"`
	Description      string                 `json:"description"`
	NetSurplusTotal  float64                `json:"net_surplus_total"`
	Distribution     []NetSurplusAllocation `json:"distribution" gorm:"-"` // Alokasi distribusi
	Members          []NetSurplusMember     `json:"members" gorm:"-"`      // Alokasi distribusi
	DistributionData string                 `gorm:"type:JSON"`
	MemberData       string                 `gorm:"type:JSON"`
	ProfitLossData   string                 `gorm:"type:JSON"`
	ProfitLoss       ProfitLossReport       `gorm:"-"`
	Transactions     []TransactionModel     `json:"transactions" gorm:"-"`
	Status           string                 `json:"status"`
	SavingsTotal     float64
	LoanTotal        float64
	TransactionTotal float64
}

func (NetSurplusModel) TableName() string {
	return "net_surplus"
}

func (n *NetSurplusModel) BeforeCreate(tx *gorm.DB) error {
	if n.ID == "" {
		n.ID = uuid.New().String()

	}
	n.DistributionData = "[]"
	n.MemberData = "[]"
	n.ProfitLossData = "{}"
	n.Status = "DRAFT"
	return nil
}

type NetSurplusAllocation struct {
	Key              string  `json:"key"`
	Name             string  `json:"name"`       // Nama alokasi (misalnya, "Simpanan Anggota", "Dana Sosial")
	Percentage       float64 `json:"percentage"` // Persentase alokasi
	Amount           float64 `json:"amount"`     // Jumlah alokasi (dihitung berdasarkan persentase)
	AccountID        *string `json:"account_id"`
	AccountCashID    *string `json:"account_cash_id"`
	AccountExpenseID *string `json:"account_expense_id"`
	Balance          float64 `json:"balance"`
}

type NetSurplusMember struct {
	ID                                   string  `json:"id"`
	FullName                             string  `json:"full_name"`
	MemberID                             string  `json:"member_id"`
	SavingsTotal                         float64 `json:"savings_total"`
	LoanTotal                            float64 `json:"loan_total"`
	TransactionTotal                     float64 `json:"transaction_total"`
	NetSurplusMandatorySavingsAllocation float64 `json:"net_surplus_mandatory_savings_allocation"`
	NetSurplusBusinessProfitAllocation   float64 `json:"net_surplus_business_profit_allocation"`
	Status                               string  `json:"status"`
}
