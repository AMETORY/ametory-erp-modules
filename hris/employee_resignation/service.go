package employee_resignation

import (
	"net/http"
	"time"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type EmployeeResignationService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

// NewEmployeeResignationService creates a new instance of EmployeeResignationService.
//
// The service is created by providing a GORM database instance and an ERP context.
// The ERP context is used for authentication and authorization purposes, while the
// database instance is used for CRUD (Create, Read, Update, Delete) operations.
func NewEmployeeResignationService(ctx *context.ERPContext) *EmployeeResignationService {
	return &EmployeeResignationService{db: ctx.DB, ctx: ctx}
}

// Migrate migrates the database for the EmployeeResignationService.
//
// It uses GORM's AutoMigrate function to create the table for EmployeeResignation
// if it does not already exist.
//
// If the migration fails, the error is returned to the caller.
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.EmployeeResignation{},
	)
}

// CreateEmployeeResignation creates a new employee resignation record in the database.
//
// The employee resignation record is created using the provided EmployeeResignationModel
// as input, and any errors are returned to the caller.
func (e *EmployeeResignationService) CreateEmployeeResignation(employeeResignation *models.EmployeeResignation) error {
	return e.db.Create(employeeResignation).Error
}

// GetEmployeeResignationByID retrieves an employee resignation record by ID from the database.
//
// The employee resignation record is queried using the GORM First method, and any errors are
// returned to the caller. If the employee resignation is not found, a nil pointer is
// returned together with a gorm.ErrRecordNotFound error.
//
// If the retrieval is successful, the employee resignation record is returned together with
// its associated Employee, Company, and Approver (with User) models.
func (e *EmployeeResignationService) GetEmployeeResignationByID(id string) (*models.EmployeeResignation, error) {
	var employeeResignation models.EmployeeResignation
	err := e.db.
		Preload("Employee").
		Preload("Company").
		Preload("Approver.User").
		Where("id = ?", id).First(&employeeResignation).Error
	if err != nil {
		return nil, err
	}

	return &employeeResignation, nil
}

// UpdateEmployeeResignation updates an existing employee resignation record in the database.
//
// The function takes an EmployeeResignation model as input and uses it to update the
// corresponding record in the database. Any errors are returned to the caller.
func (e *EmployeeResignationService) UpdateEmployeeResignation(employeeResignation *models.EmployeeResignation) error {
	return e.db.Save(employeeResignation).Error
}

// DeleteEmployeeResignation deletes an employee resignation record from the database by ID.
//
// It takes an ID as input and returns an error if the deletion operation fails.
// The function uses GORM to delete the employee resignation data from the database.
// If the deletion is successful, the error is nil. Otherwise, the error contains
// information about what went wrong.

func (e *EmployeeResignationService) DeleteEmployeeResignation(id string) error {
	return e.db.Delete(&models.EmployeeResignation{}, "id = ?", id).Error
}

// FindAllEmployeeResignations retrieves a paginated list of employee resignations from the database.
//
// The function takes an http.Request as input and applies various filters based on the request
// parameters such as company ID, user ID, search term, date range, specific date, and approver ID.
// The results can be ordered based on a specified order parameter or defaults to ordering by
// resignation_date in descending order.
//
// It utilizes pagination to manage the result set and applies any necessary request modifications
// using the utils.FixRequest utility. The function returns a paginated page of EmployeeResignation
// and an error if the operation fails.
func (e *EmployeeResignationService) FindAllEmployeeResignations(request *http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := e.db.
		Preload("Employee").
		Preload("Company").
		Preload("Approver.User").
		Model(&models.EmployeeResignation{})
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	if request.Header.Get("ID-User") != "" {
		stmt = stmt.Where("user_id = ?", request.Header.Get("ID-User"))
	}

	if request.URL.Query().Get("search") != "" {
		stmt = stmt.Where("reason LIKE ?", "%"+request.URL.Query().Get("search")+"%")
	}
	if request.URL.Query().Get("start_date") != "" && request.URL.Query().Get("end_date") != "" {
		stmt = stmt.Where("resignation_date >= ? AND resignation_date <= ?", request.URL.Query().Get("start_date"), request.URL.Query().Get("end_date"))
	}
	if request.URL.Query().Get("date") != "" {
		stmt = stmt.Where("DATE(resignation_date) = ?", request.URL.Query().Get("date"))
	}
	if request.URL.Query().Get("approver_id") != "" {
		stmt = stmt.Where("approver_id = ?", request.URL.Query().Get("approver_id"))
	}

	if request.URL.Query().Get("order") != "" {
		stmt = stmt.Order(request.URL.Query().Get("order"))
	} else {
		stmt = stmt.Order("resignation_date DESC")
	}
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.EmployeeResignation{})
	page.Page = page.Page + 1
	return page, nil
}

// CountByEmployeeID returns a map of counts of employee resignation records
// filtered by the given employee ID and date range, grouped by status.
//
// The map contains the following keys:
// - REQUESTED: the count of records with "REQUESTED" status
// - APPROVED: the count of records with "APPROVED" status
// - REJECTED: the count of records with "REJECTED" status
//
// Parameters:
//   - employeeID: The ID of the employee whose resignation records are being counted.
//   - startDate: The start date of the date range for filtering resignation records.
//   - endDate: The end date of the date range for filtering resignation records.
//
// Returns:
//   - A map of resignation counts by status.
//   - An error if the operation fails.
func (e *EmployeeResignationService) CountByEmployeeID(employeeID string, startDate *time.Time, endDate *time.Time) (map[string]int64, error) {
	var countREQUESTED, countAPPROVED, countREJECTED int64
	counts := make(map[string]int64)
	e.db.Model(&models.EmployeeResignation{}).
		Where("employee_id = ? AND status = ? AND date >= ? AND date <= ?", employeeID, "REQUESTED", startDate, endDate).
		Count(&countREQUESTED)
	e.db.Model(&models.EmployeeResignation{}).
		Where("employee_id = ? AND status = ? AND date >= ? AND date <= ?", employeeID, "APPROVED", startDate, endDate).
		Count(&countAPPROVED)
	e.db.Model(&models.EmployeeResignation{}).
		Where("employee_id = ? AND status = ? AND date >= ? AND date <= ?", employeeID, "REJECTED", startDate, endDate).
		Count(&countREJECTED)

	counts["REQUESTED"] = countREQUESTED
	counts["APPROVED"] = countAPPROVED
	counts["REJECTED"] = countREJECTED

	return counts, nil
}
