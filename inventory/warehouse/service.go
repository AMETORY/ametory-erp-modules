package warehouse

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type WarehouseService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

// NewWarehouseService creates a new instance of WarehouseService.
//
// It initializes the service with the provided Gorm database connection and ERP
// context. The WarehouseService provides operations to manage warehouse records
// in the database.
//
// Parameters:
//   db - a pointer to the Gorm DB instance
//   ctx - a pointer to the ERPContext
//
// Returns:
//   a pointer to a newly created WarehouseService instance

func NewWarehouseService(db *gorm.DB, ctx *context.ERPContext) *WarehouseService {
	return &WarehouseService{db: db, ctx: ctx}
}

// Migrate creates the database tables for the warehouse module if they do not exist.
//
// This function is intended to be used for database migrations and should be
// called during the application's setup process.
//
// Parameters:
//
//	db - a pointer to the Gorm DB instance
//
// Returns:
//
//	an error if the migration failed
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&models.WarehouseModel{})
}

// CreateWarehouse adds a new warehouse record to the database.
//
// This method receives a pointer to a WarehouseModel containing
// the warehouse details to be added. It uses GORM to insert the
// warehouse data into the warehouses table.
//
// Returns:
//   - nil if the operation is successful.
//   - an error if the operation fails.

// CreateWarehouse adds a new warehouse record to the database.
//
// This method receives a pointer to a WarehouseModel containing
// the warehouse details to be added. It uses GORM to insert the
// warehouse data into the warehouses table.
//
// Returns:
//   - nil if the operation is successful.
//   - an error if the operation fails.
func (s *WarehouseService) CreateWarehouse(data *models.WarehouseModel) error {
	return s.db.Create(data).Error
}

// UpdateWarehouse updates an existing warehouse record in the database.
//
// It takes an ID and a pointer to a WarehouseModel as input and returns an error
// if the operation fails. The function uses GORM to update the warehouse data
// in the warehouses table where the ID matches.
//
// If the update is successful, the error is nil. Otherwise, the error contains
// information about what went wrong.
func (s *WarehouseService) UpdateWarehouse(id string, data *models.WarehouseModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

// DeleteWarehouse deletes a warehouse in the database.
//
// It takes an ID as input and attempts to delete the warehouse with the given ID
// from the database. If the deletion is successful, the function returns nil. If
// the deletion operation fails, the function returns an error.
func (s *WarehouseService) DeleteWarehouse(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.WarehouseModel{}).Error
}

// GetWarehouseByID retrieves a warehouse record from the database by its ID.
//
// It takes an ID string as input and returns a pointer to a WarehouseModel
// containing the warehouse details if the record is found. If the record is
// not found, the function returns an error.
func (s *WarehouseService) GetWarehouseByID(id string) (*models.WarehouseModel, error) {
	var invoice models.WarehouseModel
	err := s.db.Where("id = ?", id).First(&invoice).Error
	return &invoice, err
}

// GetWarehouseByCode retrieves a warehouse record from the database by its code.
//
// It takes a code string as input and returns a pointer to a WarehouseModel
// containing the warehouse details if the record is found. If the record is
// not found, the function returns an error.
func (s *WarehouseService) GetWarehouseByCode(code string) (*models.WarehouseModel, error) {
	var invoice models.WarehouseModel
	err := s.db.Where("code = ?", code).First(&invoice).Error
	return &invoice, err
}

// GetWarehouses retrieves a paginated list of warehouses from the database.
//
// It takes an HTTP request and a search query string as input. The search query
// is applied to the warehouse name, code and description fields. If the request
// contains a company ID header, the method also filters the result by the
// company ID. The function utilizes pagination to manage the result set and
// applies any necessary request modifications using the utils.FixRequest utility.
//
// The function returns a paginated page of WarehouseModel and an error if the
// operation fails.
func (s *WarehouseService) GetWarehouses(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db
	if search != "" {
		stmt = stmt.Where("warehouses.description ILIKE ? OR  warehouses.name ILIKE ? OR warehouses.code ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
			"%"+search+"%",
		)
	}
	stmt = stmt.Preload("Company")
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	stmt = stmt.Model(&models.WarehouseModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.WarehouseModel{})
	page.Page = page.Page + 1
	return page, nil
}
