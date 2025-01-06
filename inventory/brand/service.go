package brand

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type BrandService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewBrandService(db *gorm.DB, ctx *context.ERPContext) *BrandService {
	return &BrandService{db: db, ctx: ctx}
}

func (s *BrandService) CreateBrand(data *BrandModel) error {
	return s.db.Create(data).Error
}

func (s *BrandService) UpdateBrand(id string, data *BrandModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

func (s *BrandService) DeleteBrand(id string) error {
	return s.db.Where("id = ?", id).Delete(&BrandModel{}).Error
}

func (s *BrandService) GetBrandByID(id string) (*BrandModel, error) {
	var invoice BrandModel
	err := s.db.Where("id = ?", id).First(&invoice).Error
	return &invoice, err
}

func (s *BrandService) GetBrandByCode(code string) (*BrandModel, error) {
	var invoice BrandModel
	err := s.db.Where("code = ?", code).First(&invoice).Error
	return &invoice, err
}

func (s *BrandService) GetBrands(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db
	if search != "" {
		stmt = stmt.Where("brands.description ILIKE ? OR brands.name ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
		)
	}
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	request.URL.Query().Get("page")
	stmt = stmt.Model(&BrandModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]BrandModel{})
	page.Page = page.Page + 1
	return page, nil
}
