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

func NewWarehouseService(db *gorm.DB, ctx *context.ERPContext) *WarehouseService {
	return &WarehouseService{db: db, ctx: ctx}
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&models.WarehouseModel{})
}

func (s *WarehouseService) CreateWarehouse(data *models.WarehouseModel) error {
	return s.db.Create(data).Error
}

func (s *WarehouseService) UpdateWarehouse(id string, data *models.WarehouseModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

func (s *WarehouseService) DeleteWarehouse(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.WarehouseModel{}).Error
}

func (s *WarehouseService) GetWarehouseByID(id string) (*models.WarehouseModel, error) {
	var invoice models.WarehouseModel
	err := s.db.Where("id = ?", id).First(&invoice).Error
	return &invoice, err
}

func (s *WarehouseService) GetWarehouseByCode(code string) (*models.WarehouseModel, error) {
	var invoice models.WarehouseModel
	err := s.db.Where("code = ?", code).First(&invoice).Error
	return &invoice, err
}

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
