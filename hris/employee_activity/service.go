package employee_activity

import (
	"net/http"
	"strings"
	"time"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type EmployeeActivityService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

// NewEmployeeActivityService creates a new EmployeeActivityService instance.
//
// The EmployeeActivityService is a service that provides operations to manipulate
// employee activities.
//
// The service is created by providing a GORM database instance and an ERP context.
// The ERP context is used for authentication and authorization purposes, while the
// database instance is used for CRUD (Create, Read, Update, Delete) operations.
func NewEmployeeActivityService(ctx *context.ERPContext) *EmployeeActivityService {
	return &EmployeeActivityService{db: ctx.DB, ctx: ctx}
}

// Migrate runs the database migration for the EmployeeActivityModel. It creates
// the table if it does not exist and modifies it if it does exist.
//
// The function returns an error if the migration fails.
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.EmployeeActivityModel{},
	)
}

// CreateEmployeeActivity creates a new employee activity in the database.
//
// The function takes an EmployeeActivityModel as input and returns an error if the
// operation fails. The function uses GORM to create the employee activity data in
// the employee_activities table.
//
// If the creation is successful, the error is nil. Otherwise, the error contains
// information about what went wrong.
func (service *EmployeeActivityService) CreateEmployeeActivity(activity *models.EmployeeActivityModel) error {
	// utils.LogJson(activity.AssignedEmployees)
	return service.db.Create(activity).Error
}

// GetEmployeeActivityByID retrieves an employee activity by its ID from the database.
//
// The function takes an activity ID as input and returns a pointer to an EmployeeActivityModel
// and an error if the operation fails. It uses GORM to preload related models, including
// Employee, Approver, AssignedEmployees, and Attendance, ensuring all relevant data is
// fetched. If the activity is found, it returns the model; otherwise, it returns an error
// indicating what went wrong.
func (service *EmployeeActivityService) GetEmployeeActivityByID(id string) (*models.EmployeeActivityModel, error) {
	var activity models.EmployeeActivityModel
	err := service.db.
		Preload("Employee", func(tx *gorm.DB) *gorm.DB {
			return tx.Preload("User").
				Preload("JobTitle").
				Preload("WorkLocation").
				Preload("WorkShift").
				Preload("Branch")
		}).
		Preload("Approver.User").
		Preload("AssignedEmployees.User").
		Preload("Attendance").
		First(&activity, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &activity, nil
}

// UpdateEmployeeActivity updates an existing employee activity in the database.
//
// The function takes an EmployeeActivityModel pointer as input and returns an error if
// the operation fails. The function uses GORM to update the employee activity data in
// the employee_activities table.
//
// If the update is successful, the error is nil. Otherwise, the error contains
// information about what went wrong.
func (service *EmployeeActivityService) UpdateEmployeeActivity(activity *models.EmployeeActivityModel) error {
	return service.db.Save(activity).Error
}

// DeleteEmployeeActivity deletes an employee activity from the database by ID.
//
// The function takes an activity ID as input and returns an error if the deletion
// operation fails. The function uses GORM to delete the employee activity data from
// the employee_activities table.
//
// If the deletion is successful, the error is nil. Otherwise, the error contains
// information about what went wrong.
func (service *EmployeeActivityService) DeleteEmployeeActivity(id string) error {
	return service.db.Delete(&models.EmployeeActivityModel{}, "id = ?", id).Error
}

// FindAll retrieves a paginated list of employee activities.
//
// The method uses GORM to query the database for employee activities, preloading the
// associated Employee, Approver, and AssignedEmployees models. It applies a filter
// based on the start date and end date if provided in the HTTP request query.
// Additionally, it applies a filter based on the employee IDs if provided, and a
// filter based on the activity types if provided. The function utilizes pagination
// to manage the result set and applies any necessary request modifications using
// the utils.FixRequest utility.
//
// The function returns a paginated page of EmployeeActivityModel and an error if
// the operation fails.
func (service *EmployeeActivityService) FindAll(request *http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := service.db.
		Preload("Employee", func(tx *gorm.DB) *gorm.DB {
			return tx.Preload("User").Preload("JobTitle")
		}).
		Preload("Approver.User").
		Preload("AssignedEmployees.User").
		Model(&models.EmployeeActivityModel{})

	if request.URL.Query().Get("start_date") != "" {
		stmt = stmt.Where("start_date >= ?", request.URL.Query().Get("start_date"))
	}

	if request.URL.Query().Get("end_date") != "" {
		stmt = stmt.Where("start_date <= ?", request.URL.Query().Get("end_date"))
	}

	if request.URL.Query().Get("employee_ids") != "" {
		stmt = stmt.Where("employee_id in (?)", strings.Split(request.URL.Query().Get("employee_ids"), ","))
	}

	if request.URL.Query().Get("types") != "" {
		stmt = stmt.Where("activity_type in (?)", strings.Split(request.URL.Query().Get("types"), ","))
	}

	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.EmployeeActivityModel{})
	page.Page = page.Page + 1
	return page, nil
}

// FindAllByEmployeeID retrieves a list of employee activities for a given employee ID.
//
// This function takes an HTTP request object, an employee ID, and an activity type as
// parameters. It filters the result set based on the employee ID and the activity type
// if provided. Additionally, it applies filters based on the search query, start date,
// end date, and date if provided in the HTTP request query. The function utilizes
// pagination to manage the result set and applies any necessary request modifications
// using the utils.FixRequest utility.
//
// The function returns a paginated page of EmployeeActivityModel and an error if the
// operation fails.
func (service *EmployeeActivityService) FindAllByEmployeeID(request *http.Request, employeeID string, activityType string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := service.db.Model(&models.EmployeeActivityModel{}).Where("employee_id = ?", employeeID)
	if activityType != "" {
		stmt = stmt.Where("activity_type = ?", activityType)
	}

	if request.URL.Query().Get("search") != "" {
		stmt = stmt.Where("name LIKE ?", "%"+request.URL.Query().Get("search")+"%")
	}

	if request.URL.Query().Get("start_date") != "" && request.URL.Query().Get("end_date") != "" {
		stmt = stmt.Where("start_date >= ? AND end_date <= ?", request.URL.Query().Get("start_date"), request.URL.Query().Get("end_date"))
	} else if request.URL.Query().Get("start_date") != "" {
		stmt = stmt.Where("start_date = ?", request.URL.Query().Get("start_date"))
	}

	if request.URL.Query().Get("date") != "" {
		stmt = stmt.Where("start_date = ?", request.URL.Query().Get("date"))
	}

	if request.URL.Query().Get("order") != "" {
		stmt = stmt.Order(request.URL.Query().Get("order"))
	} else {
		stmt = stmt.Order("start_time DESC")
	}

	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.EmployeeActivityModel{})
	page.Page = page.Page + 1
	return page, nil
}

// FindAssignmentByEmployeeID retrieves a paginated list of assignments for a specific employee.
//
// This function queries the database to find tasks assigned to the provided employee ID,
// using the activity_assigned_employees table to join records. It filters the results based
// on the search query, start date, end date, and order specified in the HTTP request query
// parameters. The function uses pagination to manage the result set and applies any necessary
// request modifications using the utils.FixRequest utility.
//
// Parameters:
//
//	request (*http.Request): The HTTP request containing query parameters for filtering and ordering.
//	employeeID (string): The ID of the employee whose assignments are to be retrieved.
//
// Returns:
//
//	paginate.Page: A paginated page containing the employee's assignments.
//	error: An error object if the operation fails, or nil if it is successful.
func (service *EmployeeActivityService) FindAssignmentByEmployeeID(request *http.Request, employeeID string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := service.db.Model(&models.EmployeeActivityModel{}).
		Joins("JOIN activity_assigned_employees ON activity_assigned_employees.employee_activity_model_id = employee_activities.id").
		Where("activity_assigned_employees.employee_model_id = ?", employeeID)
	stmt = stmt.Where("activity_type = ?", "TASK")
	if request.URL.Query().Get("search") != "" {
		stmt = stmt.Where("name LIKE ?", "%"+request.URL.Query().Get("search")+"%")
	}

	if request.URL.Query().Get("start_date") != "" && request.URL.Query().Get("end_date") != "" {
		stmt = stmt.Where("start_date >= ? AND end_date <= ?", request.URL.Query().Get("start_date"), request.URL.Query().Get("end_date"))
	} else if request.URL.Query().Get("start_date") != "" {
		stmt = stmt.Where("start_date = ?", request.URL.Query().Get("start_date"))
	}

	if request.URL.Query().Get("date") != "" {
		stmt = stmt.Where("start_date = ?", request.URL.Query().Get("date"))
	}

	if request.URL.Query().Get("order") != "" {
		stmt = stmt.Order(request.URL.Query().Get("order"))
	} else {
		stmt = stmt.Order("start_time DESC")
	}

	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.EmployeeActivityModel{})
	page.Page = page.Page + 1
	return page, nil
}

// FindApprovalByEmployeeID retrieves a paginated list of employee activities that require approval by the given employee.
//
// The function takes an HTTP request object and an employee ID as parameters. It uses the query parameters in the request
// to filter and order the activities. The activities are sorted by start time in descending order by default, but the order
// can be changed by specifying the "order" query parameter.
//
// The function returns a Page object containing the activities and the pagination information. The Page object contains the
// following fields:
//
//	Records: []models.EmployeeActivityModel
//	Page: int
//	PageSize: int
//	TotalPages: int
//	TotalRecords: int
//
// If the operation is not successful, the function returns an error object.
func (service *EmployeeActivityService) FindApprovalByEmployeeID(request *http.Request, employeeID string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := service.db.Model(&models.EmployeeActivityModel{})

	stmt = stmt.Where("approver_id = ?", employeeID)

	if request.URL.Query().Get("search") != "" {
		stmt = stmt.Where("name LIKE ?", "%"+request.URL.Query().Get("search")+"%")
	}

	if request.URL.Query().Get("start_date") != "" && request.URL.Query().Get("end_date") != "" {
		stmt = stmt.Where("start_date >= ? AND end_date <= ?", request.URL.Query().Get("start_date"), request.URL.Query().Get("end_date"))
	} else if request.URL.Query().Get("start_date") != "" {
		stmt = stmt.Where("start_date = ?", request.URL.Query().Get("start_date"))
	}

	if request.URL.Query().Get("date") != "" {
		stmt = stmt.Where("start_date = ?", request.URL.Query().Get("date"))
	}

	if request.URL.Query().Get("order") != "" {
		stmt = stmt.Order(request.URL.Query().Get("order"))
	} else {
		stmt = stmt.Order("start_time DESC")
	}

	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.EmployeeActivityModel{})
	page.Page = page.Page + 1
	return page, nil
}

// GetActivitySummaryByEmployeeID retrieves the summary of employee activities by the given employee ID and date.
//
// The function takes an employee ID and a date as parameters and returns a map containing the number of tasks,
// assignments, approvals, and visits for the given employee on the given date. The map keys are "TASK", "ASSIGNMENT",
// "APPROVAL", and "VISIT".
//
// If the operation is not successful, the function returns an error object.
func (service *EmployeeActivityService) GetActivitySummaryByEmployeeID(employeeID string, date time.Time) (map[string]int64, error) {
	var summary = make(map[string]int64)
	taskCount := int64(0)
	taskAssignmentCount := int64(0)
	taskApprovalCount := int64(0)
	taskVisitCount := int64(0)
	service.db.Model(&models.EmployeeActivityModel{}).
		Where("employee_id = ?", employeeID).
		Where("activity_type = ?", "TASK").
		Where("DATE(start_date) = ?", date).
		Count(&taskCount)
	service.db.Model(&models.EmployeeActivityModel{}).
		Where("employee_id = ?", employeeID).
		Where("activity_type = ?", "VISIT").
		Where("DATE(start_date) = ?", date).
		Count(&taskVisitCount)
	service.db.Model(&models.EmployeeActivityModel{}).
		Where("approver_id = ?", employeeID).
		Where("activity_type = ?", "TASK").
		Where("DATE(start_date) = ?", date).
		Count(&taskApprovalCount)
	service.db.Model(&models.EmployeeActivityModel{}).
		Joins("JOIN activity_assigned_employees ON activity_assigned_employees.employee_activity_model_id = employee_activities.id").
		Where("activity_assigned_employees.employee_model_id = ?", employeeID).
		Where("activity_type = ?", "TASK").
		Where("DATE(start_date) = ?", date).
		Count(&taskAssignmentCount)
	summary["ASSIGNMENT"] = taskAssignmentCount
	summary["TASK"] = taskCount
	summary["APPROVAL"] = taskApprovalCount
	summary["VISIT"] = taskApprovalCount
	return summary, nil
}
