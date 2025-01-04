package warehouse

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
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

func (s *WarehouseService) CreateWarehouse(data *WarehouseModel) error {
	return s.db.Create(data).Error
}

func (s *WarehouseService) UpdateWarehouse(id string, data *WarehouseModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

func (s *WarehouseService) DeleteWarehouse(id string) error {
	return s.db.Where("id = ?", id).Delete(&WarehouseModel{}).Error
}

func (s *WarehouseService) GetWarehouseByID(id string) (*WarehouseModel, error) {
	var invoice WarehouseModel
	err := s.db.Where("id = ?", id).First(&invoice).Error
	return &invoice, err
}

func (s *WarehouseService) GetWarehouseByCode(code string) (*WarehouseModel, error) {
	var invoice WarehouseModel
	err := s.db.Where("code = ?", code).First(&invoice).Error
	return &invoice, err
}

func (s *WarehouseService) GetWarehouses(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db
	if search != "" {
		stmt = stmt.Where("warehouses.description LIKE ? OR  warehouses.name LIKE ? OR warehouses.code LIKE ?",
			"%"+search+"%",
			"%"+search+"%",
			"%"+search+"%",
		)
	}
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	stmt = stmt.Model(&WarehouseModel{})
	page := pg.With(stmt).Request(request).Response(&[]WarehouseModel{})
	return page, nil
}
