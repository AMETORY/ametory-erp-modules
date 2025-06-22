package attendance

import (
	"errors"
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/hris/attendance_policy"
	"github.com/AMETORY/ametory-erp-modules/hris/employee"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type AttendanceService struct {
	db                      *gorm.DB
	ctx                     *context.ERPContext
	attendancePolicyService *attendance_policy.AttendancePolicyService
	employeeService         *employee.EmployeeService
}

func NewAttendanceService(ctx *context.ERPContext, employeeService *employee.EmployeeService, attendancePolicyService *attendance_policy.AttendancePolicyService) *AttendanceService {
	return &AttendanceService{
		db:                      ctx.DB,
		ctx:                     ctx,
		employeeService:         employeeService,
		attendancePolicyService: attendancePolicyService,
	}
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.AttendanceModel{},
	)
}

func (a *AttendanceService) Create(m *models.AttendanceModel) error {
	return a.db.Create(m).Error
}

func (a *AttendanceService) FindOne(id string) (*models.AttendanceModel, error) {
	m := &models.AttendanceModel{}
	if err := a.db.Where("id = ?", id).First(m).Error; err != nil {
		return nil, err
	}

	return m, nil
}

func (a *AttendanceService) FindAll(request *http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := a.db.Model(&models.AttendanceModel{})
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.AttendanceModel{})
	page.Page = page.Page + 1
	return page, nil
}

func (a *AttendanceService) Update(id string, m *models.AttendanceModel) error {
	return a.db.Where("id = ?", id).Updates(m).Error
}

func (a *AttendanceService) Delete(id string) error {
	return a.db.Where("id = ?", id).Delete(&models.AttendanceModel{}).Error
}

func (a *AttendanceService) GetEligiblePolicy(employeeID string) (*models.AttendancePolicy, error) {
	emp, err := a.employeeService.GetEmployeeByID(employeeID)
	if err != nil {
		return nil, err
	}

	policy, err := a.attendancePolicyService.FindBestPolicy(*emp.CompanyID, emp.BranchID, emp.OrganizationID, emp.WorkShiftID)
	if err != nil {
		return nil, err
	}

	return policy, nil
}

func (a *AttendanceService) CreateAttendance(m models.AttendanceCheckInput) (*models.AttendanceModel, error) {
	if m.EmployeeID == nil {
		return nil, errors.New("employee id is required")
	}
	employee, err := a.employeeService.GetEmployeeByID(*m.EmployeeID)
	if err != nil {
		return nil, err
	}

	policy, err := a.GetEligiblePolicy(employee.ID)
	if err != nil {
		return nil, err
	}

	status := a.EvaluateAttendance(policy, m)
	var attendance models.AttendanceModel

	if m.IsClockIn {
		attendance.ID = utils.Uuid()
		attendance.EmployeeID = &employee.ID
		attendance.CompanyID = employee.CompanyID
		attendance.BranchID = employee.BranchID
		attendance.OrganizationID = employee.OrganizationID
		attendance.WorkShiftID = employee.WorkShiftID
		attendance.Status = string(status)
		attendance.ClockIn = m.Now
		attendance.ClockInLat = m.Lat
		attendance.ClockInLng = m.Lng

		err := a.Create(&attendance)
		if err != nil {
			return nil, err
		}

		// TODO: Check if employee is on leave

	} else {
		att, err := a.FindOne(*m.AttendanceID)
		if err != nil {
			return nil, err
		}
		attendance = *att
		attendance.ClockOut = &m.Now
		attendance.ClockOutLat = m.Lat
		attendance.ClockOutLng = m.Lng
		attendance.Status = string(status)
		err = a.Update(attendance.ID, &attendance)
		if err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func (a *AttendanceService) EvaluateAttendance(policy *models.AttendancePolicy, input models.AttendanceCheckInput) models.AttendanceStatus {
	// Default to Accept
	status := models.Active

	// 1. Lokasi
	if policy.LocationEnabled && policy.Lat != nil && policy.Lng != nil && policy.MaxAttendanceRadius != nil {
		distance := utils.CalculateDistance(*policy.Lat, *policy.Lng, *input.Lat, *input.Lng)
		if distance > *policy.MaxAttendanceRadius {
			return policy.OnLocationFailure
		}
	}

	// 2. Face detection
	if policy.OnFaceNotDetected != "" && !input.IsFaceDetected {
		return policy.OnFaceNotDetected
	}

	// 3. Waktu & Toleransi
	if input.IsClockIn {
		status = evaluateClockIn(policy, input)
	} else {
		status = evaluateClockOut(policy, input)
	}

	return status
}

func evaluateClockIn(policy *models.AttendancePolicy, input models.AttendanceCheckInput) models.AttendanceStatus {
	actual := input.Now
	scheduled := input.ScheduledClockIn

	early := scheduled.Add(-policy.EarlyInToleranceInTime)
	late := scheduled.Add(policy.LateInToleranceInTime)

	switch {
	case actual.Before(early):
		return policy.OnEarlyInFailure
	case actual.After(late):
		return policy.OnClockInFailure
	default:
		return models.Active
	}
}

func evaluateClockOut(policy *models.AttendancePolicy, input models.AttendanceCheckInput) models.AttendanceStatus {
	actual := input.Now
	scheduled := input.ScheduledClockOut

	early := scheduled.Add(-policy.EarlyOutToleranceInTime)
	late := scheduled.Add(policy.LateOutToleranceInTime)

	switch {
	case actual.Before(early):
		return policy.OnEarlyOutFailure
	case actual.After(late):
		return policy.OnClockOutFailure
	default:
		return models.Active
	}
}
