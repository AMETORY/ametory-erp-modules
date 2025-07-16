package employee_overtime

import (
	"net/http"
	"time"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type EmployeeOvertimeService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

// NewEmployeeOvertimeService creates a new instance of EmployeeOvertimeService.
//
// The EmployeeOvertimeService is responsible for managing operations related to employee overtime.
// This function initializes the service with a GORM database and an ERP context, which are used
// for database operations and authentication/authorization purposes respectively.
//
// Parameters:
//   - ctx: A pointer to an ERPContext that provides access to the database and other context-specific
//     information.
//
// Returns:
//   - A pointer to a new instance of EmployeeOvertimeService.
func NewEmployeeOvertimeService(ctx *context.ERPContext) *EmployeeOvertimeService {
	return &EmployeeOvertimeService{db: ctx.DB, ctx: ctx}
}

// Migrate migrates the database for the EmployeeOvertimeService.
//
// It uses GORM's AutoMigrate function to create the tables for EmployeeOvertimeModel
// if they do not already exist. If the migration fails, the error is returned to the caller.
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.EmployeeOvertimeModel{},
	)
}

// CreateEmployeeOvertime creates a new employee overtime record in the database.
//
// This function takes an EmployeeOvertimeModel as input and attempts to persist
// it to the database. It returns an error if the operation fails.
//
// Parameters:
//  - employeeOvertime: A pointer to EmployeeOvertimeModel containing the details
//    of the overtime to be created.
//
// Returns:
//  - error: An error object if the creation fails, otherwise nil.

func (e *EmployeeOvertimeService) CreateEmployeeOvertime(employeeOvertime *models.EmployeeOvertimeModel) error {
	return e.db.Create(employeeOvertime).Error
}

// GetEmployeeOvertimeByID retrieves an employee overtime by its ID from the database.
//
// The function takes an overtime ID as input and returns a pointer to an EmployeeOvertimeModel
// and an error if the operation fails. It uses GORM to preload related models, including
// Employee, Approver, Attendance, and ApprovalByAdmin, ensuring all relevant data is
// fetched. If the overtime is found, it returns the model; otherwise, it returns an error
// indicating what went wrong.
func (e *EmployeeOvertimeService) GetEmployeeOvertimeByID(id string) (*models.EmployeeOvertimeModel, error) {
	var employeeOvertime models.EmployeeOvertimeModel
	err := e.db.
		Preload("Employee", func(tx *gorm.DB) *gorm.DB {
			return tx.Preload("User").
				Preload("JobTitle").
				Preload("WorkLocation").
				Preload("WorkShift").
				Preload("Branch")
		}).
		Preload("Approver.User").
		Preload("Attendance").
		Preload("ApprovalByAdmin").
		Where("id = ?", id).First(&employeeOvertime).Error
	if err != nil {
		return nil, err
	}
	return &employeeOvertime, nil
}

// UpdateEmployeeOvertime updates an existing employee overtime record in the database.
//
// The function takes an EmployeeOvertimeModel pointer as input and returns an error if
// the operation fails. The function uses GORM to update the employee overtime data in
// the employee_overtimes table where the ID matches.
//
// If the update is successful, the error is nil. Otherwise, the error contains
// information about what went wrong.
func (e *EmployeeOvertimeService) UpdateEmployeeOvertime(employeeOvertime *models.EmployeeOvertimeModel) error {
	return e.db.Save(employeeOvertime).Error
}

// DeleteEmployeeOvertime deletes an employee overtime record from the database by ID.
//
// The function takes an ID as input and returns an error if the deletion operation fails.
// The function uses GORM to delete the employee overtime data from the employee_overtimes table.
// If the deletion is successful, the error is nil. Otherwise, the error contains
// information about what went wrong.
func (e *EmployeeOvertimeService) DeleteEmployeeOvertime(id string) error {
	return e.db.Delete(&models.EmployeeOvertimeModel{}, "id = ?", id).Error
}

// FindAllByEmployeeID retrieves a paginated list of employee overtime records by employee ID.
//
// This function takes an HTTP request and an employee ID as inputs. It uses GORM to query
// the database for employee overtime records, preloading the associated Employee (with User,
// JobTitle, WorkLocation, WorkShift, Branch) and Approver (with User) models. It applies
// various filters based on the request parameters such as company ID, search term, date range,
// specific date, approver ID, and reviewer ID. The results can be ordered based on a specified
// order parameter or defaults to ordering by start_time_request in descending order.
//
// Pagination is utilized to manage the result set, and any necessary request modifications
// are applied using the utils.FixRequest utility. The function returns a paginated page of
// EmployeeOvertimeModel and an error if the operation fails.
func (e *EmployeeOvertimeService) FindAllByEmployeeID(request *http.Request, employeeID string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := e.db.
		Preload("Employee", func(tx *gorm.DB) *gorm.DB {
			return tx.Preload("User").
				Preload("JobTitle").
				Preload("WorkLocation").
				Preload("WorkShift").
				Preload("Branch")
		}).
		Preload("Approver.User").
		Where("employee_id = ?", employeeID).
		Model(&models.EmployeeOvertimeModel{})
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}

	if request.URL.Query().Get("search") != "" {
		stmt = stmt.Where("reason LIKE ?", "%"+request.URL.Query().Get("search")+"%")
	}
	if request.URL.Query().Get("start_date") != "" && request.URL.Query().Get("end_date") != "" {
		stmt = stmt.Where("start_time_request >= ? AND end_time_request <= ?", request.URL.Query().Get("start_date"), request.URL.Query().Get("end_date"))
	} else if request.URL.Query().Get("start_date") != "" {
		stmt = stmt.Where("start_time_request = ?", request.URL.Query().Get("start_date"))
	}
	if request.URL.Query().Get("date") != "" {
		stmt = stmt.Where("DATE(start_time_request) = ?", request.URL.Query().Get("date"))
	}
	if request.URL.Query().Get("approver_id") != "" {
		stmt = stmt.Where("approver_id = ?", request.URL.Query().Get("approver_id"))
	}
	if request.URL.Query().Get("reviewer_id") != "" {
		stmt = stmt.Where("reviewer_id = ?", request.URL.Query().Get("reviewer_id"))
	}
	if request.URL.Query().Get("order") != "" {
		stmt = stmt.Order(request.URL.Query().Get("order"))
	} else {
		stmt = stmt.Order("start_time_request DESC")
	}
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.EmployeeOvertimeModel{})
	page.Page = page.Page + 1
	return page, nil
}

// FindAllEmployeeOvertimes retrieves a paginated list of employee overtime records.
//
// This function takes an HTTP request as input and returns a paginated page of
// EmployeeOvertimeModel and an error if the operation fails. It uses GORM to query
// the database for employee overtime records, preloading the associated Employee (with User,
// JobTitle, WorkLocation, WorkShift, Branch) and Approver (with User) models. It applies
// various filters based on the request parameters such as company ID, search term, date range,
// specific date, approver ID, and reviewer ID. The results can be ordered based on a specified
// order parameter or defaults to ordering by start_time_request in descending order.
//
// Pagination is utilized to manage the result set, and any necessary request modifications
// are applied using the utils.FixRequest utility.
func (e *EmployeeOvertimeService) FindAllEmployeeOvertimes(request *http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := e.db.
		Preload("Employee", func(tx *gorm.DB) *gorm.DB {
			return tx.Preload("User").
				Preload("JobTitle").
				Preload("WorkLocation").
				Preload("WorkShift").
				Preload("Branch")
		}).
		Preload("Approver.User").
		Model(&models.EmployeeOvertimeModel{})
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}

	if request.URL.Query().Get("search") != "" {
		stmt = stmt.Where("reason LIKE ?", "%"+request.URL.Query().Get("search")+"%")
	}
	if request.URL.Query().Get("start_date") != "" && request.URL.Query().Get("end_date") != "" {
		stmt = stmt.Where("start_time_request >= ? AND end_time_request <= ?", request.URL.Query().Get("start_date"), request.URL.Query().Get("end_date"))
	} else if request.URL.Query().Get("start_date") != "" {
		stmt = stmt.Where("start_time_request = ?", request.URL.Query().Get("start_date"))
	}
	if request.URL.Query().Get("date") != "" {
		stmt = stmt.Where("DATE(start_time_request) = ?", request.URL.Query().Get("date"))
	}
	if request.URL.Query().Get("approver_id") != "" {
		stmt = stmt.Where("approver_id = ?", request.URL.Query().Get("approver_id"))
	}
	if request.URL.Query().Get("reviewer_id") != "" {
		stmt = stmt.Where("reviewer_id = ?", request.URL.Query().Get("reviewer_id"))
	}
	if request.URL.Query().Get("order") != "" {
		stmt = stmt.Order(request.URL.Query().Get("order"))
	} else {
		stmt = stmt.Order("start_time_request DESC")
	}
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.EmployeeOvertimeModel{})
	page.Page = page.Page + 1
	return page, nil
}

// CountByEmployeeID returns a map of counts of employee overtime records
// filtered by the given employee ID and date range, grouped by status.
//
// The map contains the following keys:
// - PENDING: the count of records with "PENDING" status
// - APPROVED: the count of records with "APPROVED" status
// - REJECTED: the count of records with "REJECTED" status
//
// Parameters:
//   - employeeID: The ID of the employee whose overtime records are being counted.
//   - startDate: The start date of the date range for filtering overtime records.
//   - endDate: The end date of the date range for filtering overtime records.
//
// Returns:
//   - A map of overtime counts by status.
//   - An error if the operation fails.
func (e *EmployeeOvertimeService) CountByEmployeeID(employeeID string, startDate *time.Time, endDate *time.Time) (map[string]int64, error) {
	var countPENDING, countAPPROVED, countREJECTED int64
	counts := make(map[string]int64)
	e.db.Model(&models.EmployeeOvertimeModel{}).
		Where("employee_id = ? AND status = ? AND start_time_request >= ? AND end_time_request <= ?", employeeID, "PENDING", startDate, endDate).
		Count(&countPENDING)
	e.db.Model(&models.EmployeeOvertimeModel{}).
		Where("employee_id = ? AND status = ? AND start_time_request >= ? AND end_time_request <= ?", employeeID, "APPROVED", startDate, endDate).
		Count(&countAPPROVED)
	e.db.Model(&models.EmployeeOvertimeModel{}).
		Where("employee_id = ? AND status = ? AND start_time_request >= ? AND end_time_request <= ?", employeeID, "REJECTED", startDate, endDate).
		Count(&countREJECTED)

	counts["PENDING"] = countPENDING
	counts["APPROVED"] = countAPPROVED
	counts["REJECTED"] = countREJECTED

	return counts, nil
}
