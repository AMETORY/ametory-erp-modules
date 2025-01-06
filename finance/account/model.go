package account

import (
	"fmt"

	"github.com/AMETORY/ametory-erp-modules/company"
	"github.com/AMETORY/ametory-erp-modules/utils"
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
	utils.BaseModel
	Name                  string                `json:"name" bson:"name"`
	Code                  string                `json:"code" bson:"code"`
	Color                 string                `json:"color" bson:"color"`
	Description           string                `json:"description" bson:"description"`
	IsDeletable           bool                  `json:"is_deletable" bson:"is_deletable"`
	IsReport              bool                  `json:"is_report" bson:"is_report" gorm:"-"`
	IsAccountReport       bool                  `json:"is_account_report" bson:"is_account_report" gorm:"-"`
	IsCashflowReport      bool                  `json:"is_cashflow_report" bson:"is_cashflow_report" gorm:"-"`
	IsPDF                 bool                  `json:"is_pdf" bson:"is_pdf" gorm:"-"`
	Type                  AccountType           `json:"type" bson:"type"`
	Category              string                `json:"category" bson:"category"`
	CashflowGroup         string                `json:"cashflow_group" bson:"cashflow_group"`
	CashflowSubGroup      string                `json:"cashflow_subgroup" bson:"cashflow_group"`
	IsTax                 bool                  `json:"is_tax" bson:"is_tax" gorm:"default:false"`
	TypeLabel             string                `gorm:"-" json:"type_label"`
	CashflowGroupLabel    string                `gorm:"-" json:"cashflow_group_label"`
	CashflowSubGroupLabel string                `gorm:"-" json:"cashflow_subgroup_label"`
	CompanyID             *string               `json:"company_id"`
	Company               *company.CompanyModel `gorm:"foreignKey:CompanyID"`
	TransactionCount      int64                 `gorm:"-" json:"transaction_count"`
	Balance               float64               `gorm:"-" json:"balance"`
	BalanceBefore         float64               `gorm:"-" json:"balance_before"`
	HasOpeningBalance     bool                  `gorm:"-" json:"has_opening_balance"`
	// Transactions          []Transaction `gorm:"-"`
}

// Migrate menjalankan migrasi database untuk model account
func Migrate(db *gorm.DB) error {
	fmt.Println("Migrating account model...")
	return db.AutoMigrate(&AccountModel{})
}

func (u *AccountModel) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}

func (AccountModel) TableName() string {
	return "accounts"
}
