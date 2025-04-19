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
		&models.WarehouseLocationModel{},
	)
}

func (s *StorageService) CreateWarehouse(data *models.WarehouseModel, lat, lng float64) error {
	if err := s.db.Create(data).Error; err != nil {
		return err
	}

	loc := models.WarehouseLocationModel{
		Name:        data.Name,
		Type:        "WAREHOUSE",
		WarehouseID: &data.ID,
		Address:     data.Address,
		Latitude:    lat,
		Longitude:   lng,
	}
	return s.CreateLocation(&loc)
}

func (s *StorageService) CreateLocation(data *models.WarehouseLocationModel) error {
	if err := s.db.Create(data).Error; err != nil {
		return err
	}
	return nil
}

func (s *StorageService) UpdateWarehouseLocation(id string, data *models.WarehouseLocationModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

func (s *StorageService) DeleteWarehouseLocation(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.WarehouseLocationModel{}).Error
}

func (s *StorageService) GetWarehouseLocationByID(id string) (*models.WarehouseLocationModel, error) {
	var invoice models.WarehouseLocationModel
	err := s.db.Where("id = ?", id).First(&invoice).Error
	return &invoice, err
}

func (s *StorageService) GetWarehouseLocations(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db
	if search != "" {
		stmt = stmt.Where("name ILIKE ?",
			"%"+search+"%",
		)
	}
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ? or company_id is null", request.Header.Get("ID-Company"))
	}
	request.URL.Query().Get("page")
	stmt = stmt.Model(&models.WarehouseLocationModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.WarehouseLocationModel{})
	page.Page = page.Page + 1
	return page, nil
}
