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

func NewStorageService(db *gorm.DB, ctx *context.ERPContext, inventoryService *inventory.InventoryService) *StorageService {
	return &StorageService{db: db, ctx: ctx, inventoryService: inventoryService}
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.LocationPointModel{},
	)
}

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

func (s *StorageService) UpdateWarehouse(id string, data *models.WarehouseModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

func (s *StorageService) DeleteWarehouse(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.WarehouseModel{}).Error
}
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

func (s *StorageService) UpdateWarehouseLocation(id string, data *models.LocationPointModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

func (s *StorageService) DeleteWarehouseLocation(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.LocationPointModel{}).Error
}

func (s *StorageService) GetWarehouseLocationByID(id string) (*models.LocationPointModel, error) {
	var location models.LocationPointModel
	err := s.db.Preload("Warehouse").
		Preload("Province").
		Preload("Regency").
		Preload("District").
		Preload("Village").Where("id = ?", id).First(&location).Error
	return &location, err
}
func (s *StorageService) GetWarehouseLocations(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.Preload("Warehouse").
		Preload("Province").
		Preload("Regency").
		Preload("District").
		Preload("Village")
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
