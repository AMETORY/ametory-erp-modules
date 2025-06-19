package hris

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/hris/attendance"
	"github.com/AMETORY/ametory-erp-modules/hris/attendance_policy"
	"github.com/AMETORY/ametory-erp-modules/hris/deduction_setting"
	"github.com/AMETORY/ametory-erp-modules/hris/employee"
	"github.com/AMETORY/ametory-erp-modules/hris/employee_activity"
	"github.com/AMETORY/ametory-erp-modules/hris/employee_overtime"
	"github.com/AMETORY/ametory-erp-modules/hris/payroll"
)

type HRISservice struct {
	ctx                     *context.ERPContext
	AttendanceService       *attendance.AttendanceService
	AttendancePolicyService *attendance_policy.AttendancePolicyService
	DeductionSettingService *deduction_setting.DeductionSettingService
	EmployeeActivityService *employee_activity.EmployeeActivityService
	EmployeeService         *employee.EmployeeService
	EmployeeOvertimeService *employee_overtime.EmployeeOvertimeService
	PayrollService          *payroll.PayrollService
}

func NewHRISservice(ctx *context.ERPContext) *HRISservice {
	service := HRISservice{
		ctx:                     ctx,
		AttendanceService:       attendance.NewAttendanceService(ctx),
		AttendancePolicyService: attendance_policy.NewAttendancePolicyService(ctx),
		DeductionSettingService: deduction_setting.NewDeductionSettingService(ctx),
		EmployeeActivityService: employee_activity.NewEmployeeActivityService(ctx),
		EmployeeService:         employee.NewEmployeeService(ctx),
		EmployeeOvertimeService: employee_overtime.NewEmployeeOvertimeService(ctx),
		PayrollService:          payroll.NewPayrollService(ctx),
	}
	if !service.ctx.SkipMigration {
		service.Migrate()
	}
	return &service
}

func (s *HRISservice) Migrate() error {
	if s.ctx.SkipMigration {
		return nil
	}
	if err := attendance.Migrate(s.ctx.DB); err != nil {
		return err
	}
	if err := attendance_policy.Migrate(s.ctx.DB); err != nil {
		return err
	}
	if err := deduction_setting.Migrate(s.ctx.DB); err != nil {
		return err
	}
	if err := employee_activity.Migrate(s.ctx.DB); err != nil {
		return err
	}
	if err := employee.Migrate(s.ctx.DB); err != nil {
		return err
	}
	if err := employee_overtime.Migrate(s.ctx.DB); err != nil {
		return err
	}
	if err := payroll.Migrate(s.ctx.DB); err != nil {
		return err
	}
	return nil
}
