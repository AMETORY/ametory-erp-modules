package work_shift

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/hris/employee"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type WorkShiftService struct {
	db              *gorm.DB
	ctx             *context.ERPContext
	employeeService *employee.EmployeeService
}

// NewWorkShiftService creates a new instance of WorkShiftService.
//
// The service is initialized with an ERP context and an EmployeeService.
// The ERP context is used for authentication and authorization, while
// the EmployeeService is used to manage employee-related operations.
func NewWorkShiftService(ctx *context.ERPContext, employeeService *employee.EmployeeService) *WorkShiftService {
	return &WorkShiftService{db: ctx.DB, ctx: ctx, employeeService: employeeService}
}

// Migrate migrates the database for the WorkShiftService.
//
// It uses GORM's AutoMigrate function to create the table for WorkShiftModel
// if it does not already exist. If the migration fails, the error is returned to the caller.
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.WorkShiftModel{},
	)
}

// CreateWorkShift creates a new work shift record in the database.
//
// The function takes a WorkShiftModel pointer as input and attempts to insert
// a new record into the work_shifts table using GORM. If the insertion is successful,
// it returns nil; otherwise, it returns an error indicating the cause of the failure.
func (a *WorkShiftService) CreateWorkShift(m *models.WorkShiftModel) error {
	return a.db.Create(m).Error
}

// FindWorkShiftByID retrieves a work shift record by ID from the database.
//
// The function takes a work shift ID as input and uses it to query the database for the work shift record.
// The associated Employee model is preloaded.
//
// The function returns the work shift record and an error if the operation fails. If the work shift record is not found,
// a nil pointer is returned together with a gorm.ErrRecordNotFound error.
func (a *WorkShiftService) FindWorkShiftByID(id string) (*models.WorkShiftModel, error) {
	m := &models.WorkShiftModel{}
	if err := a.db.Where("id = ?", id).First(m).Error; err != nil {
		return nil, err
	}

	return m, nil
}

// FindAllWorkShift retrieves a paginated list of work shift records from the database.
//
// It takes an http.Request as input and applies various filters based on the request
// parameters such as company ID, search term, date range, and employee IDs. The results can be
// ordered based on a specified order parameter or defaults to ordering by start_date in descending
// order.
//
// The function returns a paginated page of WorkShiftModel and an error if the operation fails.
func (a *WorkShiftService) FindAllWorkShift(request *http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := a.db.Model(&models.WorkShiftModel{})
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.WorkShiftModel{})
	page.Page = page.Page + 1
	return page, nil
}

// UpdateWorkShift updates a work shift record in the database.
//
// The function takes a work shift ID and a WorkShiftModel pointer as input and attempts to update
// the work shift record with the given ID in the work_shifts table using GORM. If the update
// is successful, it returns nil; otherwise, it returns an error indicating the cause of the failure.
func (a *WorkShiftService) UpdateWorkShift(id string, m *models.WorkShiftModel) error {
	return a.db.Where("id = ?", id).Updates(m).Error
}

// DeleteWorkShift deletes a work shift record from the database.
//
// It takes a work shift ID as input and attempts to delete the work shift
// with the given ID from the database. If the deletion is successful, the
// function returns nil; otherwise, it returns an error indicating the cause
// of the failure.
func (a *WorkShiftService) DeleteWorkShift(id string) error {
	return a.db.Where("id = ?", id).Delete(&models.WorkShiftModel{}).Error
}
