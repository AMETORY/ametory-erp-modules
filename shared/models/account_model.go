package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AccountType adalah enum untuk tipe account
type AccountType string

const (
	ASSET            AccountType = "ASSET"
	LIABILITY        AccountType = "LIABILITY"
	EQUITY           AccountType = "EQUITY"
	REVENUE          AccountType = "REVENUE"
	INCOME           AccountType = "INCOME"
	EXPENSE          AccountType = "EXPENSE"
	COST             AccountType = "COST"
	PAYABLE          AccountType = "PAYABLE"
	RECEIVABLE       AccountType = "RECEIVABLE"
	CONTRA_ASSET     AccountType = "CONTRA_ASSET"
	CONTRA_LIABILITY AccountType = "CONTRA_LIABILITY"
	CONTRA_EQUITY    AccountType = "CONTRA_EQUITY"
	CONTRA_REVENUE   AccountType = "CONTRA_REVENUE"
	CONTRA_EXPENSE   AccountType = "CONTRA_EXPENSE"
)

// AccountModel adalah model database untuk account
type AccountModel struct {
	shared.BaseModel
	Name                  string        `json:"name"`
	Code                  string        `json:"code"`
	Color                 string        `json:"color"`
	Description           string        `json:"description"`
	IsDeletable           bool          `json:"is_deletable"`
	IsReport              bool          `json:"is_report" gorm:"-"`
	IsAccountReport       bool          `json:"is_account_report" gorm:"-"`
	IsCashflowReport      bool          `json:"is_cashflow_report" gorm:"-"`
	IsPDF                 bool          `json:"is_pdf" gorm:"-"`
	Type                  AccountType   `json:"type"`
	Category              string        `json:"category"`
	CashflowGroup         string        `json:"cashflow_group"`
	CashflowSubGroup      string        `json:"cashflow_subgroup"`
	IsTax                 bool          `json:"is_tax" gorm:"default:false"`
	TypeLabel             string        `gorm:"-" json:"type_label,omitempty"`
	CashflowGroupLabel    string        `gorm:"-" json:"cashflow_group_label,omitempty"`
	CashflowSubGroupLabel string        `gorm:"-" json:"cashflow_subgroup_label,omitempty"`
	CompanyID             *string       `json:"company_id,omitempty"`
	Company               *CompanyModel `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE;" json:"company,omitempty"`
	UserID                *string       `json:"user_id,omitempty"`
	User                  *UserModel    `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;" json:"user,omitempty"`
	TransactionCount      int64         `gorm:"-" json:"transaction_count,omitempty"`
	Balance               float64       `gorm:"-" json:"balance,omitempty"`
	BalanceBefore         float64       `gorm:"-" json:"balance_before,omitempty"`
	HasOpeningBalance     bool          `gorm:"-" json:"has_opening_balance,omitempty"`
	IsInventoryAccount    bool          `json:"is_inventory_account,omitempty"`
	IsCogsAccount         bool          `json:"is_cogs_account,omitempty"`
	IsDiscount            bool          `json:"is_discount,omitempty"`
	IsReturn              bool          `json:"is_return,omitempty"`
	IsProfitLossAccount   bool          `json:"is_profit_loss_account,omitempty"`
	// Transactions          []Transaction `gorm:"constraint:OnDelete:CASCADE;"`
}

// Migrate menjalankan migrasi database untuk model account

func (u *AccountModel) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}

func (AccountModel) TableName() string {
	return "accounts"
}

type AccountReport struct {
	Account        AccountModel       `json:"account"`
	TotalBalance   float64            `json:"total_balance"`
	BalanceBefore  float64            `json:"balance_before"`
	CurrentBalance float64            `json:"current_balance"`
	StartDate      *time.Time         `json:"start_date"`
	EndDate        *time.Time         `json:"end_date"`
	Transactions   []TransactionModel `json:"transactions"`
}
