package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PayRollModel struct {
	shared.BaseModel
	PayRollPeriodeID                *string              `json:"pay_roll_periode_id"`
	PayRollPeriode                  *PayRollPeriodeModel `gorm:"foreignKey:PayRollPeriodeID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"pay_roll_periode"`
	PayRollNumber                   string               `json:"pay_roll_number"`
	Title                           string               `json:"title"`
	Notes                           string               `json:"notes"`
	StartDate                       time.Time            `json:"start_date" binding:"required"`
	EndDate                         time.Time            `json:"end_date" binding:"required"`
	Files                           string               `json:"files" gorm:"default:'[]'"`
	TotalIncome                     float64              `json:"total_income"`
	TotalReimbursement              float64              `json:"total_reimbursement"`
	TotalDeduction                  float64              `json:"total_deduction"`
	TotalTax                        float64              `json:"total_tax"`
	TaxCost                         float64              `json:"tax_cost"`
	NetIncome                       float64              `json:"net_income"`
	NetIncomeBeforeTaxCost          float64              `json:"net_income_before_tax_cost"`
	TakeHomePay                     float64              `json:"take_home_pay"`
	TotalPayable                    float64              `json:"total_payable"`
	TaxAllowance                    float64              `json:"tax_allowance"`
	TaxTariff                       float64              `json:"tax_tariff"`
	IsGrossUp                       bool                 `json:"is_gross_up"`
	IsEffectiveRateAverage          bool                 `json:"is_effective_rate_average"`
	Status                          string               `json:"status" gorm:"type:varchar(20);default:'DRAFT'"` //'DRAFT', 'RUNNING', 'FINISHED'
	Attachments                     []string             `json:"attachments" gorm:"-"`
	Transactions                    []TransactionModel   `json:"transactions" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	PayableTransactions             []TransactionModel   `json:"payable_transactions" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Items                           []PayrollItemModel   `json:"items" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Costs                           []PayRollCostModel   `json:"costs" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	TakeHomePayCounted              string               `json:"take_home_pay_counted" gorm:"-"`
	TakeHomePayReimbursementCounted string               `json:"take_home_pay_reimbursement_counted" gorm:"-"`
	TaxPaymentID                    *string              `json:"tax_payment_id"`
	EmployeeID                      string               `binding:"required" json:"employee_id"`
	Employee                        EmployeeModel        `gorm:"foreignKey:EmployeeID" `
	TaxSummary                      CountTaxSummary      `gorm:"-" json:"tax_summary"`
	// BpjsSetting                     *thirdparty.Bpjs     `gorm:"-" `
	PayRollReportItemID *string      `json:"pay_roll_report_item_id" gorm:"type:char(36)"`
	CompanyID           string       `json:"company_id" gorm:"not null"`
	Company             CompanyModel `gorm:"foreignKey:CompanyID"`
	IsLocked            bool         `json:"is_locked"`
}

func (p *PayRollModel) TableName() string {
	return "pay_rolls"
}

func (p *PayRollModel) BeforeCreate(tx *gorm.DB) (err error) {

	if p.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}

type CountTaxSummary struct {
	JobExpenseMonthly               float64 `json:"job_expense_monthly"`
	JobExpenseYearly                float64 `json:"job_expense_yearly"`
	PtkpYearly                      float64 `json:"ptkp_yearly"`
	GrossIncomeMonthly              float64 `json:"gross_income_monthly"`
	GrossIncomeYearly               float64 `json:"gross_income_yearly"`
	PkpMonthly                      float64 `json:"pkp_monthly"`
	PkpYearly                       float64 `json:"pkp_yearly"`
	TaxYearlyBasedOnProgressiveRate float64 `json:"tax_yearly_based_on_progressive_rate"`
	TaxYearly                       float64 `json:"tax_yearly"`
	TaxMonthly                      float64 `json:"tax_monthly"`
	NetIncomeMonthly                float64 `json:"net_income_monthly"`
	NetIncomeYearly                 float64 `json:"net_income_yearly"`
	CutoffPensiunMonthly            float64 `json:"cutoff_pensiun_monthly"`
	CutoffPensiunYearly             float64 `json:"cutoff_pensiun_yearly"`
	CutoffMonthly                   float64 `json:"cutoff_monthly"`
	CutoffYearly                    float64 `json:"cutoff_yearly"`
	Ter                             float64 `json:"ter"`
}
