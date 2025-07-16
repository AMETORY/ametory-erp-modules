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

// NewAttendanceService returns a new instance of AttendanceService.
//
// The AttendanceService is a service that provides operations to manipulate
// attendance.
//
// The service is created by providing a GORM database instance, an ERP context,
// an employee service, and an attendance policy service. The ERP context is used
// for authentication and authorization purposes, while the database instance is
// used for CRUD (Create, Read, Update, Delete) operations. The employee service
// and the attendance policy service are used to fetch related data.
func NewAttendanceService(ctx *context.ERPContext, employeeService *employee.EmployeeService, attendancePolicyService *attendance_policy.AttendancePolicyService) *AttendanceService {
	return &AttendanceService{
		db:                      ctx.DB,
		ctx:                     ctx,
		employeeService:         employeeService,
		attendancePolicyService: attendancePolicyService,
	}
}

// Migrate runs the auto migration for the attendance model.
//
// The attendance model is the model that stores the attendance data of the
// employees. This function is used to create the attendance table in the
// database if it does not exist.
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.AttendanceModel{},
	)
}

// Create adds a new attendance record to the database.
//
// The function takes an AttendanceModel object as input and attempts to
// insert it into the database. If the operation is successful, it returns
// nil; otherwise, it returns an error indicating what went wrong.
//
// Parameters:
//   m (*models.AttendanceModel): The attendance model instance to be added.
//
// Returns:
//   error: An error object if the operation fails, or nil if it is successful.

func (a *AttendanceService) Create(m *models.AttendanceModel) error {
	return a.db.Create(m).Error
}

// FindOne finds an attendance record by its ID.
//
// The function takes an attendance ID as parameter and attempts to find the
// attendance record in the database. If the record is found, the function
// returns the attendance model instance; otherwise, it returns an error
// indicating what went wrong.
//
// Parameters:
//
//	id (string): The ID of the attendance record to be found.
//
// Returns:
//
//	*models.AttendanceModel: The attendance model instance if found, or nil if
//	  not found.
//	error: An error object if the operation fails, or nil if it is successful.
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

// FindAttendanceByEmployeeAndDate retrieves an attendance record for a specific employee on a given date.
//
// This function queries the database to find an attendance entry associated with the provided employee ID
// and date. It preloads related data such as employee details, attendance policies, and schedule information.
// If a matching record is found, the function returns the attendance model instance; otherwise, it returns an error.
//
// Parameters:
//   employeeID (string): The ID of the employee whose attendance record is to be retrieved.
//   date (time.Time): The date for which the attendance record is to be retrieved.
//
// Returns:
//   *models.AttendanceModel: The attendance model instance if found, or nil if not found.
//   error: An error object if the operation fails, or nil if it is successful.

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

// FindAll retrieves all attendance records from the database.
//
// The function takes an HTTP request object as parameter and uses the query
// parameters to filter the attendance records. The records are sorted by
// clock-in time in descending order by default, but the order can be changed
// by specifying the "order" query parameter.
//
// The function returns a Page object containing the attendance records and
// the pagination information. The Page object contains the following fields:
//
//	Records: []models.AttendanceModel
//	Page: int
//	PageSize: int
//	TotalPages: int
//	TotalRecords: int
//
// If the operation is not successful, the function returns an error object.
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

// Update updates an existing attendance record in the database.
//
// The function takes an attendance ID and a pointer to an AttendanceModel as
// parameters. It attempts to update the attendance record in the database
// with the provided information. If the operation is successful, it returns
// an error object indicating what went wrong.
func (a *AttendanceService) Update(id string, m *models.AttendanceModel) error {
	return a.db.Where("id = ?", id).Updates(m).Error
}

// Delete deletes an existing attendance record from the database.
//
// The function takes an attendance ID as parameter and attempts to delete the
// attendance record with the given ID from the database. If the operation is
// successful, it returns an error object indicating what went wrong.
func (a *AttendanceService) Delete(id string) error {
	return a.db.Where("id = ?", id).Delete(&models.AttendanceModel{}).Error
}

// GetEligiblePolicy retrieves the best suitable attendance policy for a given employee.
//
// This function fetches the employee details using the provided employee ID
// and subsequently uses the employee's company, branch, organization, and work shift
// information to find the most appropriate attendance policy.
//
// Parameters:
//
//	employeeID (string): The ID of the employee for whom the eligible attendance policy is to be found.
//
// Returns:
//
//	*models.AttendancePolicy: The most suitable attendance policy for the employee.
//	error: An error object if the operation fails, or nil if it is successful.
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

// GetFiles retrieves the clock-in and clock-out files associated with an attendance record.
//
// This function takes an attendance model as parameter and populates the ClockInFile and
// ClockInPicture fields with the corresponding file information. If the attendance record does
// not have any associated clock-in or clock-out files, the respective fields are not populated.
//
// Parameters:
//
//	attendance *models.AttendanceModel: The attendance record for which the clock-in and clock-out files are to be retrieved.
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

// CreateAttendance creates an attendance record based on the provided information.
//
// The function takes an attendance check input model as parameter and creates a new
// attendance record in the database. If the attendance is a clock-in, the function
// will create a new attendance record with the provided information. If the attendance
// is a clock-out, the function will update the existing attendance record with the
// provided information.
//
// The function returns the created attendance model if the operation is successful,
// or an error object if the operation fails.
//
// Parameters:
//
//	m models.AttendanceCheckInput: The attendance check input model containing the
//		employee ID, scheduled clock-in and clock-out times, notes, and pictures.
//
// Returns:
//
//	*models.AttendanceModel: The created attendance model if the operation is successful,
//		or nil if the operation fails.
//	error: An error object if the operation fails, or nil if it is successful.
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

// EvaluateAttendance checks if the attendance is valid based on the attendance policy.
// If the attendance is invalid, it will return the status and remarks based on the policy.
// The policy evaluation is done in the following order:
//  1. Location: If the location is enabled and the user is outside the maximum attendance radius,
//     the policy's OnLocationFailure status and LocationDistanceProblem remarks will be returned.
//  2. Face detection: If the face is not detected and the policy has a failure status for this,
//     the policy's OnFaceNotDetected status and FaceProblem remarks will be returned.
//  3. Time & Tolerance: The function will check if the attendance is a clock-in or clock-out and
//     call the corresponding evaluation function. If the evaluation function returns an invalid
//     status, the policy's failure status and remarks will be returned.
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

// evaluateClockIn evaluates the clock-in input based on the attendance policy.
// The function first calculates the early and late times based on the scheduled
// clock-in time and the policy's early and late tolerances. It then checks if
// the actual time is before the early time or after the late time. If so, it
// returns the corresponding failure status and remarks. If the actual time is
// within the allowed range, it returns the active status and empty remarks.
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

// evaluateClockOut evaluates the clock-out input based on the attendance policy.
// The function first calculates the early and late times based on the scheduled
// clock-out time and the policy's early and late tolerances. It then checks if
// the actual time is before the early time or after the late time. If so, it
// returns the corresponding failure status and remarks. If the actual time is
// within the allowed range, it returns the active status and empty remarks.
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
