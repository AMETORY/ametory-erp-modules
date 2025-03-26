package models

import "github.com/AMETORY/ametory-erp-modules/shared"

type CooperativeSettingModel struct {
	shared.BaseModel
	CompanyID                           *string       `gorm:"size:30" json:"company_id"`
	Company                             CompanyModel  `gorm:"foreignKey:CompanyID;references:ID" json:"company"`
	PrincipalSavingsAmount              float64       `json:"principal_savings_amount"`
	MandatorySavingsAmount              float64       `json:"mandatory_savings_amount"`
	VoluntarySavingsAmount              float64       `json:"voluntary_savings_amount"`
	NetSurplusMandatorySavings          float64       `json:"net_surplus_mandatory_savings"`
	NetSurplusReserve                   float64       `json:"net_surplus_reserve"`
	NetSurplusBusinessProfit            float64       `json:"net_surplus_business_profit"`
	NetSurplusSocialFund                float64       `json:"net_surplus_social_fund"`
	NetSurplusEducationFund             float64       `json:"net_surplus_education_fund"`
	NetSurplusManagement                float64       `json:"net_surplus_management"`
	NetSurplusOtherFunds                float64       `json:"net_surplus_other_funds"`
	PrincipalSavingsAccountID           *string       `json:"principal_savings_account_id"`
	MandatorySavingsAccountID           *string       `json:"mandatory_savings_account_id"`
	VoluntarySavingsAccountID           *string       `json:"voluntary_savings_account_id"`
	LoanAccountID                       *string       `json:"loan_account_id"`
	LoanAccountIncomeID                 *string       `json:"loan_account_income_id"`
	LoanAccountAdminFeeID               *string       `json:"loan_account_admin_fee_id"`
	NetSurplusReserveAccountID          *string       `json:"net_surplus_reserve_account_id"`
	NetSurplusMandatorySavingsAccountID *string       `json:"net_surplus_mandatory_savings_account_id"`
	NetSurplusBusinessProfitAccountID   *string       `json:"net_surplus_business_profit_account_id"`
	NetSurplusSocialFundAccountID       *string       `json:"net_surplus_social_fund_account_id"`
	NetSurplusEducationFundAccountID    *string       `json:"net_surplus_education_fund_account_id"`
	NetSurplusManagementAccountID       *string       `json:"net_surplus_management_account_id"`
	NetSurplusOtherFundsAccountID       *string       `json:"net_surplus_other_funds_account_id"`
	PrincipalSavingsAccount             *AccountModel `json:"principal_savings_account,omitempty" gorm:"foreignKey:PrincipalSavingsAccountID"`
	MandatorySavingsAccount             *AccountModel `json:"mandatory_savings_account,omitempty" gorm:"foreignKey:MandatorySavingsAccountID"`
	VoluntarySavingsAccount             *AccountModel `json:"voluntary_savings_account,omitempty" gorm:"foreignKey:VoluntarySavingsAccountID"`
	LoanAccount                         *AccountModel `json:"loan_account,omitempty" gorm:"foreignKey:LoanAccountID"`
	LoanAccountAdminFee                 *AccountModel `json:"loan_account_admin_fee,omitempty" gorm:"foreignKey:LoanAccountAdminFeeID"`
	LoanAccountIncome                   *AccountModel `json:"loan_account_income,omitempty" gorm:"foreignKey:LoanAccountIncomeID"`
	NetSurplusReserveAccount            *AccountModel `json:"net_surplus_reserve_account,omitempty" gorm:"foreignKey:NetSurplusReserveAccountID"`
	NetSurplusMandatorySavingsAccount   *AccountModel `json:"net_surplus_mandatory_savings_account,omitempty" gorm:"foreignKey:NetSurplusMandatorySavingsAccountID"`
	NetSurplusBusinessProfitAccount     *AccountModel `json:"net_surplus_business_profit_account,omitempty" gorm:"foreignKey:NetSurplusBusinessProfitAccountID"`
	NetSurplusSocialFundAccount         *AccountModel `json:"net_surplus_social_fund_account,omitempty" gorm:"foreignKey:NetSurplusSocialFundAccountID"`
	NetSurplusEducationFundAccount      *AccountModel `json:"net_surplus_education_fund_account,omitempty" gorm:"foreignKey:NetSurplusEducationFundAccountID"`
	NetSurplusManagementAccount         *AccountModel `json:"net_surplus_management_account,omitempty" gorm:"foreignKey:NetSurplusManagementAccountID"`
	NetSurplusOtherFundsAccount         *AccountModel `json:"net_surplus_other_funds_account,omitempty" gorm:"foreignKey:NetSurplusOtherFundsAccountID"`
	TermCondition                       string        `json:"term_condition" gorm:"type:LONGTEXT"`
	StaticCharacter                     string        `json:"static_character"`
	NumberFormat                        string        `json:"number_format"`
	AutoNumericLength                   int           `json:"auto_numeric_length"`
	RandomNumericLength                 int           `json:"random_numeric_length"`
	RandomCharacterLength               int           `json:"random_character_length"`
	InterestRatePerMonth                float64       `json:"interest_rate_per_month"`
	ExpectedProfitRatePerMonth          float64       `json:"expected_profit_rate_per_month"`
}
