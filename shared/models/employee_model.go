package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EmployeeModel struct {
	shared.BaseModel
	Email                         string                  `json:"email,omitempty"`
	FirstName                     string                  `json:"first_name,omitempty"`
	MiddleName                    string                  `json:"middle_name,omitempty"`
	LastName                      string                  `json:"last_name,omitempty"`
	Username                      string                  `json:"username,omitempty"`
	Phone                         string                  `json:"phone,omitempty"`
	JobTitleID                    *string                 `json:"job_title_id,omitempty"`
	JobTitle                      *JobTitleModel          `gorm:"foreignKey:JobTitleID;constraint:OnDelete:CASCADE" json:"job_title,omitempty"`
	BranchID                      *string                 `json:"branch_id,omitempty"`
	Branch                        BranchModel             `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	Grade                         string                  `json:"grade,omitempty"`
	Address                       string                  `json:"address,omitempty"`
	Picture                       *string                 `json:"picture,omitempty"`
	Cover                         string                  `json:"cover,omitempty"`
	StartedWork                   *time.Time              `json:"started_work,omitempty"`
	DateOfBirth                   *time.Time              `json:"date_of_birth,omitempty"`
	EmployeeIdentityNumber        string                  `json:"employee_identity_number,omitempty"`
	EmployeeCode                  string                  `json:"employee_code,omitempty"`
	FullName                      string                  `json:"full_name,omitempty"`
	ConnectedTo                   *string                 `json:"connected_to,omitempty"`
	Flag                          bool                    `json:"flag,omitempty"`
	BasicSalary                   float64                 `json:"basic_salary,omitempty"`
	PositionalAllowance           float64                 `json:"positional_allowance,omitempty"`
	TransportAllowance            float64                 `json:"transport_allowance,omitempty"`
	MealAllowance                 float64                 `json:"meal_allowance,omitempty"`
	NonTaxableIncomeLevelCode     string                  `json:"non_taxable_income_level_code,omitempty"`
	PayRolls                      []PayRollModel          `json:"pay_rolls,omitempty" gorm:"foreignKey:EmployeeID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Reimbursements                []ReimbursementModel    `json:"reimbursements,omitempty" gorm:"foreignKey:EmployeeID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	TaxPayerNumber                string                  `json:"tax_payer_number,omitempty"`
	Gender                        string                  `json:"gender,omitempty"`
	Attendance                    []AttendanceModel       `json:"attendance,omitempty" gorm:"foreignKey:EmployeeID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Organization                  *OrganizationModel      `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	OrganizationID                *string                 `json:"organization_id,omitempty"`
	WorkingType                   string                  `json:"working_type,omitempty" gorm:"default:'FULL_TIME'"`
	SalaryType                    string                  `json:"salary_type,omitempty" gorm:"default:'MONTHLY'"`
	Schedules                     []*ScheduleModel        `json:"schedules,omitempty" gorm:"many2many:schedule_employees;"`
	TotalWorkingDays              float64                 `json:"total_working_days,omitempty"`
	TotalWorkingHours             float64                 `json:"total_working_hours,omitempty"`
	DailyWorkingHours             float64                 `json:"daily_working_hours,omitempty"`
	WorkSafetyRisks               string                  `gorm:"default:'very_low'" json:"work_safety_risks,omitempty"`
	Sales                         []SalesModel            `json:"sales,omitempty" gorm:"foreignKey:EmployeeID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	IndonesianIdentityCardNumber  string                  `json:"indonesian_identity_card_number,omitempty"`
	BankAccountNumber             string                  `json:"bank_account_number,omitempty"`
	BankID                        *string                 `json:"bank_id,omitempty"`
	Bank                          *BankModel              `gorm:"foreignKey:BankID;constraint:OnDelete:CASCADE" json:"bank,omitempty"`
	CompanyID                     *string                 `json:"company_id,omitempty" gorm:"size:36;not null"`
	Company                       *CompanyModel           `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE" json:"company,omitempty"`
	UserID                        *string                 `gorm:"size:36" json:"user_id,omitempty"`
	User                          *UserModel              `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	WorkShiftID                   *string                 `json:"work_shift_id"`
	WorkShift                     *WorkShiftModel         `gorm:"foreignKey:WorkShiftID" json:"work_shift,omitempty"`
	Status                        string                  `gorm:"default:'ACTIVE'" json:"status,omitempty"`
	EmployeePushNotifTokens       []PushTokenModel        `json:"push_notification_tokens,omitempty" gorm:"foreignKey:EmployeeID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	AnnualLeaveDays               int                     `json:"annual_leave_days,omitempty"`
	LateDeductionSettingID        *string                 `json:"late_deduction_setting_id,omitempty"`
	LateDeductionSetting          *DeductionSettingModel  `gorm:"foreignKey:LateDeductionSettingID" json:"late_deduction_setting,omitempty"`
	NotPresenceDeductionSettingID *string                 `json:"not_presence_deduction_setting_id,omitempty"`
	NotPresenceDeductionSetting   *DeductionSettingModel  `gorm:"foreignKey:NotPresenceDeductionSettingID" json:"not_presence_deduction_setting,omitempty"`
	Loans                         []LoanModel             `json:"loans,omitempty" gorm:"foreignKey:EmployeeID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	IsNewUser                     bool                    `json:"is_new_user,omitempty" gorm:"-"`
	Activities                    []EmployeeActivityModel `json:"activities,omitempty" gorm:"foreignKey:EmployeeID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
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
