package hris

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/hris/attendance"
	"github.com/AMETORY/ametory-erp-modules/hris/attendance_policy"
	"github.com/AMETORY/ametory-erp-modules/hris/deduction_setting"
	"github.com/AMETORY/ametory-erp-modules/hris/employee"
	"github.com/AMETORY/ametory-erp-modules/hris/employee_activity"
	"github.com/AMETORY/ametory-erp-modules/hris/employee_business_trip"
	"github.com/AMETORY/ametory-erp-modules/hris/employee_cash_advance"
	"github.com/AMETORY/ametory-erp-modules/hris/employee_loan"
	"github.com/AMETORY/ametory-erp-modules/hris/employee_overtime"
	"github.com/AMETORY/ametory-erp-modules/hris/employee_resignation"
	"github.com/AMETORY/ametory-erp-modules/hris/leave"
	"github.com/AMETORY/ametory-erp-modules/hris/payroll"
	"github.com/AMETORY/ametory-erp-modules/hris/reimbursement"
	"github.com/AMETORY/ametory-erp-modules/hris/schedule"
	"github.com/AMETORY/ametory-erp-modules/hris/work_shift"
)

type HRISservice struct {
	ctx                         *context.ERPContext
	AttendanceService           *attendance.AttendanceService
	AttendancePolicyService     *attendance_policy.AttendancePolicyService
	DeductionSettingService     *deduction_setting.DeductionSettingService
	EmployeeActivityService     *employee_activity.EmployeeActivityService
	EmployeeService             *employee.EmployeeService
	EmployeeOvertimeService     *employee_overtime.EmployeeOvertimeService
	EmployeeCashAdvanceService  *employee_cash_advance.EmployeeCashAdvanceService
	PayrollService              *payroll.PayrollService
	LeaveService                *leave.LeaveService
	ReimbursementService        *reimbursement.ReimbursementService
	ScheduleService             *schedule.ScheduleService
	EmployeeLoanService         *employee_loan.EmployeeLoanService
	JobTitleService             *employee.JobTitleService
	WorkShiftService            *work_shift.WorkShiftService
	EmployeeBusinessTripService *employee_business_trip.EmployeeBusinessTripService
	EmployeeResignationService  *employee_resignation.EmployeeResignationService
}

// NewHRISservice creates a new instance of HRISservice.
//
// The service is created by providing a pointer to an ERPContext. The ERPContext
// is used for authentication and authorization purposes.
//
// The service is created with all the services that are a part of the HRIS module.
// The services are created with the provided ERP context and the necessary
// dependencies.
//
// The method also calls Migrate() on the service, which migrates the database
// schema to the latest version.
//
// If the SkipMigration flag is set on the ERPContext, the Migrate() method is not
// called.
//
// The service is then returned as a pointer to the caller.
func NewHRISservice(ctx *context.ERPContext) *HRISservice {
	employeeService := employee.NewEmployeeService(ctx)
	attendancePolicyService := attendance_policy.NewAttendancePolicyService(ctx)
	service := HRISservice{
		ctx:                         ctx,
		AttendanceService:           attendance.NewAttendanceService(ctx, employeeService, attendancePolicyService),
		AttendancePolicyService:     attendancePolicyService,
		DeductionSettingService:     deduction_setting.NewDeductionSettingService(ctx),
		EmployeeActivityService:     employee_activity.NewEmployeeActivityService(ctx),
		EmployeeService:             employeeService,
		EmployeeOvertimeService:     employee_overtime.NewEmployeeOvertimeService(ctx),
		EmployeeCashAdvanceService:  employee_cash_advance.NewEmployeeCashAdvanceService(ctx),
		PayrollService:              payroll.NewPayrollService(ctx, employeeService),
		LeaveService:                leave.NewLeaveService(ctx, employeeService),
		ReimbursementService:        reimbursement.NewReimbursementService(ctx, employeeService),
		ScheduleService:             schedule.NewScheduleService(ctx, employeeService),
		EmployeeLoanService:         employee_loan.NewEmployeeLoanService(ctx, employeeService),
		JobTitleService:             employee.NewJobTitleService(ctx),
		WorkShiftService:            work_shift.NewWorkShiftService(ctx, employeeService),
		EmployeeBusinessTripService: employee_business_trip.NewEmployeeBusinessTripService(ctx),
		EmployeeResignationService:  employee_resignation.NewEmployeeResignationService(ctx),
	}
	if !service.ctx.SkipMigration {
		service.Migrate()
	}
	return &service
}

// Migrate migrates the database schema to the latest version.
//
// The method calls Migrate() on all the services that are a part of the HRIS
// module. If the SkipMigration flag is set on the ERPContext, the Migrate()
// method is not called.
//
// The method returns an error if any of the services fail to migrate their
// database schema.
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
	if err := employee_cash_advance.Migrate(s.ctx.DB); err != nil {
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
	if err := employee_business_trip.Migrate(s.ctx.DB); err != nil {
		return err
	}
	if err := employee_resignation.Migrate(s.ctx.DB); err != nil {
		return err
	}
	return nil
}
