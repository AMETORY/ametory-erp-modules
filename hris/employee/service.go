package employee

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type EmployeeService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

// NewEmployeeService creates a new instance of EmployeeService.
//
// It initializes the service with a GORM database and an ERP context.
// The ERP context is used for authentication and authorization, while
// the database is used for CRUD operations on employee-related data.
func NewEmployeeService(ctx *context.ERPContext) *EmployeeService {
	return &EmployeeService{db: ctx.DB, ctx: ctx}
}

// Migrate migrates the database for the EmployeeService.
//
// It uses GORM's AutoMigrate function to create the tables for EmployeeModel
// and JobTitleModel if they do not already exist.
//
// If the migration fails, the error is returned to the caller.
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.EmployeeModel{},
		&models.JobTitleModel{},
	)
}

// CreateEmployee creates a new employee in the database.
//
// The employee is created using the provided EmployeeModel as input, and
// any errors are returned to the caller.
func (e *EmployeeService) CreateEmployee(employee *models.EmployeeModel) error {
	return e.db.Create(employee).Error
}

// GetEmployeeByID retrieves an employee by ID from the database.
//
// The employee is queried using the GORM First method, and any errors are
// returned to the caller. If the employee is not found, a nil pointer is
// returned together with a gorm.ErrRecordNotFound error.
func (e *EmployeeService) GetEmployeeByID(id string) (*models.EmployeeModel, error) {
	var employee models.EmployeeModel
	err := e.db.
		Preload("User").
		Preload("Company").
		Preload("Bank").
		Preload("JobTitle").
		Preload("Branch").
		Preload("WorkLocation").
		Preload("WorkShift").
		First(&employee, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &employee, nil
}

// UpdateEmployee updates an existing employee record in the database.
//
// It takes an EmployeeModel pointer as input, and returns an error if the
// operation fails. The function uses GORM to update the employee data in
// the employees table where the ID matches.
//
// If the update is successful, the error is nil. Otherwise, the error contains
// information about what went wrong.
func (e *EmployeeService) UpdateEmployee(employee *models.EmployeeModel) error {
	return e.db.Save(employee).Error
}

// DeleteEmployee deletes an employee record from the database by ID.
//
// It takes an ID as input and returns an error if the deletion operation fails.
// The function uses GORM to delete the employee data from the employees table.
// If the deletion is successful, the error is nil. Otherwise, the error contains
// information about what went wrong.
func (e *EmployeeService) DeleteEmployee(id string) error {
	return e.db.Delete(&models.EmployeeModel{}, id).Error
}

// FindAllEmployees retrieves a paginated list of employees.
//
// The method uses GORM to query the database for employees, preloading the
// associated User, Company, and JobTitle models. It applies a filter based on
// the company ID provided in the HTTP request header, and another filter based
// on the search parameter if provided. The function utilizes pagination to
// manage the result set and applies any necessary request modifications using
// the utils.FixRequest utility.
//
// The function returns a paginated page of EmployeeModel and an error if the
// operation fails.
func (e *EmployeeService) FindAllEmployees(request *http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := e.db.Preload("User").Preload("Company").Preload("JobTitle").Model(&models.EmployeeModel{})
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	if request.URL.Query().Get("search") != "" {
		stmt = stmt.Where("full_name ilike ? or email ilike ? or employee_identity_number ilike ? or address ilike ? or phone ilike ?",
			"%"+request.URL.Query().Get("search")+"%",
			"%"+request.URL.Query().Get("search")+"%",
			"%"+request.URL.Query().Get("search")+"%",
			"%"+request.URL.Query().Get("search")+"%",
			"%"+request.URL.Query().Get("search")+"%",
		)
	}
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.EmployeeModel{})
	page.Page = page.Page + 1
	return page, nil
}

// GetEmployeeFromUser retrieves an employee associated with a specific user ID and
// company ID from the database.
//
// It takes a user ID and a company ID as input and returns a pointer to an
// EmployeeModel and an error if the operation fails. The function uses GORM to
// query the database for an employee with the specified user ID and company ID,
// and returns the employee data if found. If the operation fails, an error is
// returned.
func (e *EmployeeService) GetEmployeeFromUser(userID string, companyID string) (*models.EmployeeModel, error) {
	var employee models.EmployeeModel
	err := e.db.
		First(&employee, "user_id = ? AND company_id = ?", userID, companyID).Error
	if err != nil {
		return nil, err
	}
	return &employee, nil
}
