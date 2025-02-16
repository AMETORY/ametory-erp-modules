package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EmployeeModel struct {
	shared.BaseModel
	Email                         string                 `json:"email"`
	FirstName                     string                 `json:"first_name"`
	MiddleName                    string                 `json:"middle_name"`
	LastName                      string                 `json:"last_name"`
	Username                      string                 `json:"username"`
	Phone                         string                 `json:"phone"`
	JobTitleID                    *string                `json:"job_title_id"`
	JobTitle                      *JobTitleModel         `gorm:"foreignKey:JobTitleID;constraint:OnDelete:CASCADE"`
	BranchID                      *string                `json:"branch_id"`
	Branch                        BranchModel            `gorm:"foreignKey:BranchID"`
	Grade                         string                 `json:"grade"`
	Address                       string                 `json:"address"`
	Picture                       *string                `json:"picture"`
	Cover                         string                 `json:"cover"`
	StartedWork                   *time.Time             `json:"started_work"`
	DateOfBirth                   *time.Time             `json:"date_of_birth"`
	EmployeeIdentityNumber        string                 `json:"employee_identity_number"`
	EmployeeCode                  string                 `json:"employee_code"`
	FullName                      string                 `json:"full_name"`
	ConnectedTo                   *string                `json:"connected_to"`
	Flag                          bool                   `json:"flag"`
	BasicSalary                   float64                `json:"basic_salary"`
	PositionalAllowance           float64                `json:"positional_allowance"`
	TransportAllowance            float64                `json:"transport_allowance"`
	MealAllowance                 float64                `json:"meal_allowance"`
	NonTaxableIncomeLevelCode     string                 `json:"non_taxable_income_level_code"`
	PayRolls                      []PayRollModel         `json:"pay_rolls" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Reimbursements                []ReimbursementModel   `json:"reimbursements" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	TaxPayerNumber                string                 `json:"tax_payer_number"`
	Gender                        string                 `json:"gender"`
	Attendance                    []AttendanceModel      `json:"attendance" gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Organization                  OrganizationModel      `gorm:"foreignKey:OrganizationID"`
	OrganizationID                *string                `json:"organization_id"`
	WorkingType                   string                 `json:"working_type" gorm:"default:'FULL_TIME'"`
	SalaryType                    string                 `json:"salary_type" gorm:"default:'MONTHLY'"`
	Schedules                     []*ScheduleModel       `json:"-" gorm:"many2many:schedule_employees;"`
	TotalWorkingDays              float64                `json:"total_working_days"`
	TotalWorkingHours             float64                `json:"total_working_hours"`
	DailyWorkingHours             float64                `json:"daily_working_hours"`
	WorkSafetyRisks               string                 `gorm:"default:'very_low'" json:"work_safety_risks"`
	Sales                         []SalesModel           `json:"sales" gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	IndonesianIdentityCardNumber  string                 `json:"indonesian_identity_card_number"`
	BankAccountNumber             string                 `json:"bank_account_number"`
	BankID                        *string                `json:"bank_id"`
	Bank                          *BankModel             `gorm:"foreignKey:BankID;constraint:OnDelete:CASCADE;"`
	CompanyID                     string                 `json:"company_id" gorm:"not null"`
	Company                       CompanyModel           `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE"`
	UserID                        *string                `json:"user_id" `
	Status                        string                 `gorm:"default:'ACTIVE'" json:"status"`
	EmployeePushNotifTokens       []PushTokenModel       `json:"push_notification_tokens" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	AnnualLeaveDays               int                    `json:"annual_leave_days" `
	Tx                            *gorm.DB               `gorm:"-"`
	LateDeductionSettingID        *string                `json:"late_deduction_setting_id"`
	LateDeductionSetting          *DeductionSettingModel `gorm:"foreignKey:LateDeductionSettingID"`
	NotPresenceDeductionSettingID *string                `json:"not_presence_deduction_setting_id"`
	NotPresenceDeductionSetting   *DeductionSettingModel `gorm:"foreignKey:NotPresenceDeductionSettingID"`
	Loans                         []LoanModel            `json:"loans" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (e EmployeeModel) TableName() string {
	return "employees"
}

func (e *EmployeeModel) BeforeCreate(tx *gorm.DB) (err error) {
	if e.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}
