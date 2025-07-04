package attendance

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

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
	if err := a.db.
		Preload("Employee", func(db *gorm.DB) *gorm.DB {
			return db.Preload("User").Preload("Branch").Preload("Organization").Preload("WorkShift").Preload("JobTitle")
		}).
		Preload("AttendancePolicy").
		Preload("ClockOutAttendancePolicy").
		Preload("Schedule").
		Where("id = ?", id).First(m).Error; err != nil {
		return nil, err
	}

	return m, nil
}

func (a *AttendanceService) FindAttendanceByEmployeeAndDate(employeeID string, date time.Time) (*models.AttendanceModel, error) {
	m := &models.AttendanceModel{}
	if err := a.db.
		Preload("Employee", func(db *gorm.DB) *gorm.DB {
			return db.Preload("User").Preload("Branch").Preload("Organization").Preload("WorkShift").Preload("JobTitle")
		}).
		Preload("AttendancePolicy").
		Preload("ClockOutAttendancePolicy").
		Preload("Schedule").
		Where("employee_id = ? AND DATE(clock_in) = ?",
			employeeID,
			date.Format("2006-01-02")).
		Order("clock_in desc").
		First(m).Error; err != nil {
		return nil, err
	}

	return m, nil
}

func (a *AttendanceService) FindAll(request *http.Request) (paginate.Page, error) {
	// fmt.Println("GET ATTENDANCES")
	pg := paginate.New()
	stmt := a.db.
		Preload("Employee.User").
		Preload("AttendancePolicy").
		Preload("ClockOutAttendancePolicy").
		Preload("Schedule").
		Model(&models.AttendanceModel{})
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	if request.URL.Query().Get("employee_ids") != "" {
		stmt = stmt.Where("employee_id IN (?)", strings.Split(request.URL.Query().Get("employee_ids"), ","))
	}
	if request.URL.Query().Get("employee_id") != "" {
		stmt = stmt.Where("employee_id IN (?)", strings.Split(request.URL.Query().Get("employee_id"), ","))
	}
	if request.URL.Query().Get("start_date") != "" {
		stmt = stmt.Where("clock_in >= ?", request.URL.Query().Get("start_date"))
	}
	if request.URL.Query().Get("end_date") != "" {
		stmt = stmt.Where("clock_in <= ?", request.URL.Query().Get("end_date"))
	}
	if request.URL.Query().Get("order") != "" {
		stmt = stmt.Order(request.URL.Query().Get("order"))
	} else {
		stmt = stmt.Order("clock_in desc")
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

func (a *AttendanceService) GetFiles(attendance *models.AttendanceModel) {
	var fileClockIn models.FileModel
	a.db.Where("ref_id = ? AND ref_type = ?", attendance.ID, "clockin").Find(&fileClockIn)
	if fileClockIn.ID != "" {
		attendance.ClockInFile = &fileClockIn
		attendance.ClockInPicture = fileClockIn.URL
	}
	var fileClockOut models.FileModel
	a.db.Where("ref_id = ? AND ref_type = ?", attendance.ID, "clockout").Find(&fileClockOut)
	if fileClockOut.ID != "" {
		attendance.ClockInFile = &fileClockOut
		attendance.ClockOutPicture = fileClockOut.URL
	}

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

	pol, err := a.attendancePolicyService.FindOne(policy.ID)
	if err != nil {
		return nil, err
	}
	policy = pol

	fmt.Println("ATTENDANCE POLICY", policy.PolicyName)

	status, remarks := a.EvaluateAttendance(policy, m)
	if status == models.Reject {
		return nil, errors.New(string(remarks))
	}
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
		attendance.Remarks = string(remarks)
		attendance.AttendancePolicyID = &policy.ID
		attendance.ScheduleID = m.ScheduleID
		attendance.ClockInNotes = m.Notes
		if status == models.Pending && remarks == models.LateInProblem {
			lateInDuration := int(m.Now.Sub(m.ScheduledClockIn).Seconds())
			attendance.LateIn = &lateInDuration
		}
		if m.File.URL != "" {
			attendance.ClockInPicture = m.File.URL
		}
		err := a.Create(&attendance)
		if err != nil {
			return nil, err
		}

		if m.File != nil {
			m.File.RefID = attendance.ID
			m.File.RefType = "clockin"
			a.db.Save(m.File)
		}

		// TODO: Check if employee is on leave

	} else {
		if m.AttendanceID == nil {
			return nil, errors.New("attendance id is required")
		}
		att, err := a.FindOne(*m.AttendanceID)
		if err != nil {
			return nil, err
		}
		fmt.Println("CLOCKIN", att.ClockIn)
		fmt.Println("CLOCKOUT", m.Now)
		workingDuration := int(m.Now.Sub(att.ClockIn).Seconds())
		fmt.Println("DURATION", workingDuration)
		attendance = *att
		attendance.ClockOut = &m.Now
		attendance.ClockOutLat = m.Lat
		attendance.ClockOutLng = m.Lng
		attendance.ClockOutRemarks = string(remarks)
		attendance.ClockOutAttendancePolicyID = &policy.ID
		attendance.ClockOutNotes = m.Notes
		attendance.WorkingDuration = &workingDuration
		attendance.Status = "DONE"
		if m.File.URL != "" {
			attendance.ClockOutPicture = m.File.URL
		}
		err = a.Update(attendance.ID, &attendance)
		if err != nil {
			return nil, err
		}

		if m.File != nil {
			m.File.RefID = attendance.ID
			m.File.RefType = "clockout"
			a.db.Save(m.File)
		}
	}

	return &attendance, nil
}

func (a *AttendanceService) EvaluateAttendance(policy *models.AttendancePolicy, input models.AttendanceCheckInput) (models.AttendanceStatus, models.Remarks) {
	// Default to Accept
	status := models.Active
	remarks := models.Empty
	// 1. Lokasi
	if policy.LocationEnabled && policy.Lat != nil && policy.Lng != nil && policy.MaxAttendanceRadius != nil {
		distance := utils.CalculateDistance(*policy.Lat, *policy.Lng, *input.Lat, *input.Lng)
		if distance > *policy.MaxAttendanceRadius {
			return policy.OnLocationFailure, models.LocationDistanceProblem
		}
	}

	// 2. Face detection
	if policy.OnFaceNotDetected != "" && !input.IsFaceDetected {
		return policy.OnFaceNotDetected, models.FaceProblem
	}

	// 3. Waktu & Toleransi
	if input.IsClockIn {
		status, remarks = evaluateClockIn(policy, input)
	} else {
		status, remarks = evaluateClockOut(policy, input)
	}

	return status, remarks
}

func evaluateClockIn(policy *models.AttendancePolicy, input models.AttendanceCheckInput) (models.AttendanceStatus, models.Remarks) {
	actual := input.Now
	scheduled := input.ScheduledClockIn

	early := scheduled.Add(-policy.EarlyInToleranceInTime * time.Minute)
	late := scheduled.Add(policy.LateInToleranceInTime * time.Minute)

	fmt.Println("\nactual", actual, "\nearly", early, "\nlate", late, "\nscheduled", scheduled)
	switch {
	case actual.Before(early):
		return policy.OnEarlyInFailure, models.EarlyInProblem
	case actual.After(late):
		return policy.OnClockInFailure, models.LateInProblem
	default:
		return models.Active, models.Empty
	}
}

func evaluateClockOut(policy *models.AttendancePolicy, input models.AttendanceCheckInput) (models.AttendanceStatus, models.Remarks) {
	actual := input.Now
	scheduled := input.ScheduledClockOut

	early := scheduled.Add(-policy.EarlyOutToleranceInTime * time.Minute)
	late := scheduled.Add(policy.LateOutToleranceInTime * time.Minute)

	fmt.Println("\nactual", actual, "\nearly", early, "\nlate", late, "\nscheduled", scheduled)
	switch {
	case actual.Before(early):
		return policy.OnEarlyOutFailure, models.EarlyOutProblem
	case actual.After(late):
		return policy.OnClockOutFailure, models.LateOutProblem
	default:
		return models.Active, models.Empty
	}
}
