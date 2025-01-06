package product

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type MasterProductService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewMasterProductService(db *gorm.DB, ctx *context.ERPContext) *MasterProductService {
	return &MasterProductService{db: db, ctx: ctx}
}

func (s *MasterProductService) CreateMasterProduct(data *MasterProductModel) error {
	return s.db.Create(data).Error
}

func (s *MasterProductService) UpdateMasterProduct(id string, data *MasterProductModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

func (s *MasterProductService) DeleteMasterProduct(id string) error {
	return s.db.Where("id = ?", id).Delete(&MasterProductModel{}).Error
}

func (s *MasterProductService) GetMasterProductByID(id string) (*MasterProductModel, error) {
	var invoice MasterProductModel
	err := s.db.Where("id = ?", id).First(&invoice).Error
	return &invoice, err
}

func (s *MasterProductService) GetMasterProductByCode(code string) (*MasterProductModel, error) {
	var invoice MasterProductModel
	err := s.db.Where("code = ?", code).First(&invoice).Error
	return &invoice, err
}

func (s *MasterProductService) GetMasterProducts(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db
	if search != "" {
		stmt = stmt.Where("master_products.description LIKE ? OR master_products.sku LIKE ? OR master_products.name LIKE ? OR master_products.barcode LIKE ?",
			"%"+search+"%",
			"%"+search+"%",
			"%"+search+"%",
			"%"+search+"%",
		)
	}
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	stmt = stmt.Model(&MasterProductModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]MasterProductModel{})
	page.Page = page.Page + 1
	return page, nil
}
