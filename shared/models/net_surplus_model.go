package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NetSurplusModel struct {
	shared.BaseModel
	CompanyID        *string                `gorm:"size:36" json:"company_id,omitempty"`
	UserID           *string                `gorm:"size:36" json:"user_id,omitempty"`
	StartDate        time.Time              `json:"start_date,omitempty"`
	EndDate          time.Time              `json:"end_date,omitempty"`
	Date             time.Time              `json:"date,omitempty"`
	Description      string                 `json:"description"`
	NetSurplusTotal  float64                `json:"net_surplus_total"`
	Distribution     []NetSurplusAllocation `json:"distribution,omitempty" gorm:"-"` // Alokasi distribusi
	Members          []NetSurplusMember     `json:"members,omitempty" gorm:"-"`      // Alokasi distribusi
	DistributionData string                 `gorm:"type:JSON" json:"distribution_data,omitempty"`
	MemberData       string                 `gorm:"type:JSON" json:"member_data,omitempty"`
	ProfitLossData   *string                `gorm:"type:JSON" json:"profit_loss_data,omitempty"`
	ProfitLoss       *ProfitLossReport      `gorm:"-" json:"profit_loss"`
	Transactions     []TransactionModel     `json:"transactions,omitempty" gorm:"-"`
	Status           string                 `json:"status"`
	SavingsTotal     float64                `json:"savings_total"`
	LoanTotal        float64                `json:"loan_total"`
	TransactionTotal float64                `json:"transaction_total"`
	NetSurplusNumber string                 `json:"net_surplus_number"`
	ClosingBookID    *string                `gorm:"size:36" json:"closing_book_id,omitempty"`
	ClosingBook      *ClosingBook           `gorm:"foreignKey:ClosingBookID;constraint:OnDelete:SET NULL" json:"closing_book,omitempty"`
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
	profitLossData := "{}"
	n.ProfitLossData = &profitLossData
	n.Status = "DRAFT"
	return nil
}

func (n *NetSurplusModel) AfterFind(tx *gorm.DB) error {
	var distributions []NetSurplusAllocation
	if err := json.Unmarshal([]byte(n.DistributionData), &distributions); err != nil {
		return err
	}
	for i, v := range distributions {
		if v.AccountID != nil {
			var account AccountModel
			if err := tx.Model(&AccountModel{}).Where("id = ?", *v.AccountID).Find(&account).Error; err != nil {
				return err
			}
			v.Account = &account
		}
		if v.AccountCashID != nil {
			var account AccountModel
			if err := tx.Model(&AccountModel{}).Where("id = ?", *v.AccountCashID).Find(&account).Error; err != nil {
				return err
			}
			v.AccountCash = &account
		}
		if v.AccountExpenseID != nil {
			var account AccountModel
			if err := tx.Model(&AccountModel{}).Where("id = ?", *v.AccountExpenseID).Find(&account).Error; err != nil {
				return err
			}
			v.AccountExpense = &account
		}
		distributions[i] = v
	}
	fmt.Println("DISTRIBUTIONS")
	utils.LogJson(distributions)
	n.Distribution = distributions

	var members []NetSurplusMember
	if err := json.Unmarshal([]byte(n.MemberData), &members); err != nil {
		return err
	}
	n.Members = members

	if n.ProfitLossData != nil {
		var profitLoss ProfitLossReport
		if err := json.Unmarshal([]byte(*n.ProfitLossData), &profitLoss); err != nil {
			return err
		}
		n.ProfitLoss = &profitLoss
	}

	n.ProfitLossData = nil
	n.MemberData = ""
	n.DistributionData = ""

	return nil
}

type NetSurplusAllocation struct {
	Key              string        `json:"key"`
	Name             string        `json:"name"`       // Nama alokasi (misalnya, "Simpanan Anggota", "Dana Sosial")
	Percentage       float64       `json:"percentage"` // Persentase alokasi
	Amount           float64       `json:"amount"`     // Jumlah alokasi (dihitung berdasarkan persentase)
	AccountID        *string       `json:"account_id"`
	Account          *AccountModel `json:"account"`
	AccountCashID    *string       `json:"account_cash_id"`
	AccountCash      *AccountModel `json:"account_cash"`
	AccountExpenseID *string       `json:"account_expense_id"`
	AccountExpense   *AccountModel `json:"account_expense"`
	Balance          float64       `json:"balance"`
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
