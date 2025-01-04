package sales

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type SalesService struct {
	ctx *context.ERPContext
	db  *gorm.DB
}

func NewSalesService(db *gorm.DB, ctx *context.ERPContext) *SalesService {
	return &SalesService{db: db, ctx: ctx}
}

func (s *SalesService) CreateSales(data *SalesModel) error {
	return s.db.Create(data).Error
}

func (s *SalesService) UpdateSales(id string, data *SalesModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

func (s *SalesService) DeleteSales(id string) error {
	return s.db.Where("id = ?", id).Delete(&SalesModel{}).Error
}

func (s *SalesService) GetSalesByID(id string) (*SalesModel, error) {
	var sales SalesModel
	err := s.db.Where("id = ?", id).First(&sales).Error
	return &sales, err
}

func (s *SalesService) GetSalesByCode(code string) (*SalesModel, error) {
	var sales SalesModel
	err := s.db.Where("code = ?", code).First(&sales).Error
	return &sales, err
}

func (s *SalesService) GetSalesBySalesNumber(salesNumber string) (*SalesModel, error) {
	var sales SalesModel
	err := s.db.Where("sales_number = ?", salesNumber).First(&sales).Error
	return &sales, err
}

func (s *SalesService) GetSales(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db
	if search != "" {
		stmt = stmt.Where("sales.description LIKE ? OR sales.code LIKE ? OR sales.sales_number LIKE ?",
			"%"+search+"%",
			"%"+search+"%",
			"%"+search+"%",
		)
	}
	stmt = stmt.Model(&SalesModel{})
	page := pg.With(stmt).Request(request).Response(&[]SalesModel{})
	return page, nil
}
