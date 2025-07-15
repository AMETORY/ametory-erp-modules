package storage

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/inventory"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type StorageService struct {
	db               *gorm.DB
	ctx              *context.ERPContext
	inventoryService *inventory.InventoryService
}

// NewStorageService creates a new instance of StorageService with the given database connection, context and inventory service.
func NewStorageService(db *gorm.DB, ctx *context.ERPContext, inventoryService *inventory.InventoryService) *StorageService {
	return &StorageService{db: db, ctx: ctx, inventoryService: inventoryService}
}

// Migrate runs the database migrations to create the necessary tables.
//
// It creates the `location_points` table.
func Migrate(db *gorm.DB) error {

	return db.AutoMigrate(
		&models.LocationPointModel{},
	)
}

// GetWarehouses retrieves a paginated list of warehouses from the database.
//
// It takes an HTTP request and a search query string as input. The method uses
// GORM to query the database for warehouses, applying the search query to the
// warehouse name, code and description fields. If the request contains a company
// ID header, the method also filters the result by the company ID. The function
// utilizes pagination to manage the result set and applies any necessary request
// modifications using the utils.FixRequest utility.
//
// The function returns a paginated page of WarehouseModel and an error if the
// operation fails.
func (s *StorageService) GetWarehouses(request http.Request, search string) (paginate.Page, error) {
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

// UpdateWarehouse updates an existing warehouse record in the database.
//
// It takes an ID and a pointer to a WarehouseModel as input and returns an error
// if the operation fails. The function uses GORM to update the warehouse data
// in the warehouses table where the ID matches.
//
// If the update is successful, the error is nil. Otherwise, the error contains
// information about what went wrong.
func (s *StorageService) UpdateWarehouse(id string, data *models.WarehouseModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

// DeleteWarehouse deletes a warehouse in the database.
//
// It takes an ID as input and attempts to delete the warehouse with the given ID
// from the database. If the deletion is successful, the function returns nil. If
// the deletion operation fails, the function returns an error.
func (s *StorageService) DeleteWarehouse(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.WarehouseModel{}).Error
}

// CreateWarehouse creates a new warehouse in the database.
//
// It takes a pointer to a WarehouseModel, latitude and longitude as input. The
// function first creates a new warehouse record in the database using the input
// data. If the latitude and longitude are not empty, it creates a new location
// point record associated with the new warehouse.
//
// The function returns an error if any of the operations fail. Otherwise, it
// returns nil.
func (s *StorageService) CreateWarehouse(data *models.WarehouseModel, lat, lng *float64) error {
	if err := s.db.Create(data).Error; err != nil {
		return err
	}
	if lat != nil && lng != nil {
		loc := models.LocationPointModel{
			Name:        data.Name,
			Type:        "WAREHOUSE",
			WarehouseID: &data.ID,
			Address:     data.Address,
			Latitude:    *lat,
			Longitude:   *lng,
		}
		return s.CreateLocation(&loc, nil)
	}
	return nil
}

// CreateLocation creates a new location point record in the database.
//
// It takes a pointer to a LocationPointModel and an optional pointer to a WarehouseModel as input.
// If the warehouse pointer is not nil, it either creates a new warehouse record in the database if
// the warehouse ID is empty or updates the existing record if the ID is not empty. The function then
// sets the WarehouseID field of the LocationPointModel to the ID of the created or updated warehouse.
// Finally, it creates a new location point record in the database using the input data.
//
// The function returns an error if any of the operations fail. Otherwise, it returns nil.
func (s *StorageService) CreateLocation(data *models.LocationPointModel, warehouse *models.WarehouseModel) error {
	if warehouse != nil {
		if warehouse.ID == "" {
			err := s.db.Create(warehouse).Error
			if err != nil {
				return err
			}
		} else {
			err := s.db.Save(warehouse).Error
			if err != nil {
				return err
			}
		}
		data.WarehouseID = &warehouse.ID
	}
	if err := s.db.Create(data).Error; err != nil {
		return err
	}

	return nil
}

// UpdateWarehouseLocation updates an existing warehouse location record in the database.
//
// It takes an ID and a pointer to a LocationPointModel as input and returns an error
// if the operation fails. The function uses GORM to update the location point data
// in the location_points table where the ID matches.
//
// If the update is successful, the error is nil. Otherwise, the error contains
// information about what went wrong.

func (s *StorageService) UpdateWarehouseLocation(id string, data *models.LocationPointModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

// DeleteWarehouseLocation deletes a warehouse location record in the database.
//
// It takes an ID as input and attempts to delete the warehouse location with the
// given ID from the database. If the deletion is successful, the function returns
// nil. If the deletion operation fails, the function returns an error.
func (s *StorageService) DeleteWarehouseLocation(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.LocationPointModel{}).Error
}

// GetWarehouseLocationByID retrieves a warehouse location record from the database by ID.
//
// It takes an ID as input and returns a pointer to a LocationPointModel and an error.
// The function uses GORM to retrieve the location point data from the location_points table
// where the ID matches and preloads the associated warehouse. If the operation fails,
// an error is returned.
func (s *StorageService) GetWarehouseLocationByID(id string) (*models.LocationPointModel, error) {
	var location models.LocationPointModel
	err := s.db.Preload("Warehouse").Where("id = ?", id).First(&location).Error
	return &location, err
}

// GetWarehouseLocations retrieves a paginated list of warehouse locations from the database.
//
// It takes an HTTP request and a search query string as input. The search query
// is applied to the name field of the location points. If the request contains
// a company ID header, the method filters results by company ID or includes
// entries with a null company ID. Pagination is applied to manage the result set,
// and any necessary request modifications are made using the utils.FixRequest utility.
//
// The function returns a paginated page of LocationPointModel and an error if the
// operation fails.

func (s *StorageService) GetWarehouseLocations(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.Preload("Warehouse")
	if search != "" {
		stmt = stmt.Where("name ILIKE ?",
			"%"+search+"%",
		)
	}
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ? or company_id is null", request.Header.Get("ID-Company"))
	}
	request.URL.Query().Get("page")
	stmt = stmt.Model(&models.LocationPointModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.LocationPointModel{})
	page.Page = page.Page + 1
	return page, nil
}
