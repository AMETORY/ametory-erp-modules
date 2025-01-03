package account

import (
	"github.com/AMETORY/ametory-erp-modules/company"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AccountModel adalah model database untuk account
type AccountModel struct {
	utils.BaseModel
	Name                  string               `json:"name" bson:"name"`
	Code                  string               `json:"code" bson:"code"`
	Color                 string               `json:"color" bson:"color"`
	Description           string               `json:"description" bson:"description"`
	IsDeletable           bool                 `json:"is_deletable" bson:"is_deletable"`
	IsReport              bool                 `json:"is_report" bson:"is_report" gorm:"-"`
	IsAccountReport       bool                 `json:"is_account_report" bson:"is_account_report" gorm:"-"`
	IsCashflowReport      bool                 `json:"is_cashflow_report" bson:"is_cashflow_report" gorm:"-"`
	IsPDF                 bool                 `json:"is_pdf" bson:"is_pdf" gorm:"-"`
	Type                  string               `json:"type" bson:"type"`
	Category              string               `json:"category" bson:"category"`
	CashflowGroup         string               `json:"cashflow_group" bson:"cashflow_group"`
	CashflowSubGroup      string               `json:"cashflow_subgroup" bson:"cashflow_group"`
	IsTax                 bool                 `json:"is_tax" bson:"is_tax" gorm:"default:false"`
	TypeLabel             string               `gorm:"-" json:"type_label"`
	CashflowGroupLabel    string               `gorm:"-" json:"cashflow_group_label"`
	CashflowSubGroupLabel string               `gorm:"-" json:"cashflow_subgroup_label" `
	CompanyID             string               `json:"company_id" gorm:"not null"`
	Company               company.CompanyModel `gorm:"foreignKey:CompanyID"`
	TransactionCount      int64                `gorm:"-"`
	Balance               float64              `gorm:"-"`
	BalanceBefore         float64              `gorm:"-"`
	HasOpeningBalance     bool                 `gorm:"-"`
	// Transactions          []Transaction `gorm:"-"`
}

// Migrate menjalankan migrasi database untuk model account
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&AccountModel{})
}

func (u *AccountModel) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}
