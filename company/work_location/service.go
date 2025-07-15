package work_location

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type WorkLocationService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

// NewWorkLocationService creates a new instance of WorkLocationService.
//
// The WorkLocationService is a service that provides operations to manipulate work
// locations.
//
// The service requires a Gorm database and an ERP context.
func NewWorkLocationService(db *gorm.DB, ctx *context.ERPContext) *WorkLocationService {
	return &WorkLocationService{db: db, ctx: ctx}
}

// CreateWorkLocation creates a new work location.
//
// The function takes a WorkLocationModel pointer and
// creates a new work location in the database.
//
// The function returns an error if the creation failed.
func (s *WorkLocationService) CreateWorkLocation(data *models.WorkLocationModel) error {
	return s.db.Create(data).Error
}

// UpdateWorkLocation updates an existing work location in the database.
//
// The function takes a work location ID and a pointer to a WorkLocationModel
// containing the data to be updated. It performs the update operation in the
// database using the provided ID as the identifier.
//
// Returns an error if the update operation fails.

func (s *WorkLocationService) UpdateWorkLocation(id string, data *models.WorkLocationModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

// DeleteWorkLocation deletes a work location in the database.
//
// The function takes a work location ID and attempts to delete the work location
// with the given ID from the database. If the deletion is successful, the
// function returns nil. If the deletion operation fails, the function returns
// an error.
func (s *WorkLocationService) DeleteWorkLocation(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.WorkLocationModel{}).Error
}

// GetWorkLocationByID retrieves a work location from the database by ID.
//
// It takes an ID as input and returns a pointer to a WorkLocationModel and an error.
// The function uses GORM to retrieve the work location data from the work_locations table.
// If the operation fails, an error is returned.
func (s *WorkLocationService) GetWorkLocationByID(id string) (*models.WorkLocationModel, error) {
	var branch models.WorkLocationModel
	if err := s.db.Where("id = ?", id).First(&branch).Error; err != nil {
		return nil, err
	}
	return &branch, nil
}

// FindAllWorkLocations retrieves a paginated list of work locations associated with a specific company.
//
// It takes an HTTP request as input and returns a paginated Page of WorkLocationModel
// and an error if the operation fails. The function applies a filter based on the
// company ID provided in the request header to further filter the work locations.

func (s *WorkLocationService) FindAllWorkLocations(request *http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.Model(&models.WorkLocationModel{})
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.WorkLocationModel{})
	page.Page = page.Page + 1
	return page, nil
}

// GetWorkLocationByEmployee retrieves a work location associated with a specific employee.
//
// It takes an EmployeeModel pointer as input and returns a pointer to a WorkLocationModel
// and an error if the operation fails. The function uses GORM to retrieve the work location
// data using the employee ID as the identifier. If the operation fails, an error is returned.
func (s *WorkLocationService) GetWorkLocationByEmployee(employee *models.EmployeeModel) (*models.WorkLocationModel, error) {
	if employee == nil {
		return nil, nil
	}

	if err := s.db.Model(&employee).Preload("WorkLocation").Find(employee).Error; err != nil {
		return nil, err
	}
	return employee.WorkLocation, nil
}
