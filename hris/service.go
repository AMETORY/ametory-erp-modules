package hris

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/hris/attendance"
	"github.com/AMETORY/ametory-erp-modules/hris/attendance_policy"
	"github.com/AMETORY/ametory-erp-modules/hris/deduction_setting"
	"github.com/AMETORY/ametory-erp-modules/hris/employee"
	"github.com/AMETORY/ametory-erp-modules/hris/employee_activity"
	"github.com/AMETORY/ametory-erp-modules/hris/employee_loan"
	"github.com/AMETORY/ametory-erp-modules/hris/employee_overtime"
	"github.com/AMETORY/ametory-erp-modules/hris/leave"
	"github.com/AMETORY/ametory-erp-modules/hris/payroll"
	"github.com/AMETORY/ametory-erp-modules/hris/reimbursement"
	"github.com/AMETORY/ametory-erp-modules/hris/schedule"
	"github.com/AMETORY/ametory-erp-modules/hris/work_shift"
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
	LeaveService            *leave.LeaveService
	ReimbursementService    *reimbursement.ReimbursementService
	ScheduleService         *schedule.ScheduleService
	EmployeeLoanService     *employee_loan.EmployeeLoanService
	JobTitleService         *employee.JobTitleService
	WorkShiftService        *work_shift.WorkShiftService
}

func NewHRISservice(ctx *context.ERPContext) *HRISservice {
	employeeService := employee.NewEmployeeService(ctx)
	service := HRISservice{
		ctx:                     ctx,
		AttendanceService:       attendance.NewAttendanceService(ctx),
		AttendancePolicyService: attendance_policy.NewAttendancePolicyService(ctx),
		DeductionSettingService: deduction_setting.NewDeductionSettingService(ctx),
		EmployeeActivityService: employee_activity.NewEmployeeActivityService(ctx),
		EmployeeService:         employeeService,
		EmployeeOvertimeService: employee_overtime.NewEmployeeOvertimeService(ctx),
		PayrollService:          payroll.NewPayrollService(ctx, employeeService),
		LeaveService:            leave.NewLeaveService(ctx, employeeService),
		ReimbursementService:    reimbursement.NewReimbursementService(ctx, employeeService),
		ScheduleService:         schedule.NewScheduleService(ctx, employeeService),
		EmployeeLoanService:     employee_loan.NewEmployeeLoanService(ctx, employeeService),
		JobTitleService:         employee.NewJobTitleService(ctx),
		WorkShiftService:        work_shift.NewWorkShiftService(ctx, employeeService),
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
	if err := leave.Migrate(s.ctx.DB); err != nil {
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
	if err := employee_loan.Migrate(s.ctx.DB); err != nil {
		return err
	}
	if err := leave.Migrate(s.ctx.DB); err != nil {
		return err
	}
	if err := schedule.Migrate(s.ctx.DB); err != nil {
		return err
	}
	if err := reimbursement.Migrate(s.ctx.DB); err != nil {
		return err
	}
	return nil
}
