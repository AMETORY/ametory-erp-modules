package product

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type PriceCategoryService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewPriceCategoryService(db *gorm.DB, ctx *context.ERPContext) *PriceCategoryService {
	return &PriceCategoryService{db: db, ctx: ctx}
}

func (s *PriceCategoryService) GetPriceCategories(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db
	if search != "" {
		stmt = stmt.Where("price_categories.name ILIKE ? OR price_categories.description ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
		)
	}

	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ? or company_id is null", request.Header.Get("ID-Company"))
	}

	stmt = stmt.Model(&models.PriceCategoryModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.PriceCategoryModel{})

	return page, nil
}

func (s *PriceCategoryService) GetPriceCategoryByID(id string) (*models.PriceCategoryModel, error) {
	var category models.PriceCategoryModel
	err := s.db.Model(&models.PriceCategoryModel{}).Where("id = ?", id).First(&category).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (s *PriceCategoryService) CreatePriceCategory(data *models.PriceCategoryModel) error {
	return s.db.Create(data).Error
}

func (s *PriceCategoryService) UpdatePriceCategory(id string, data *models.PriceCategoryModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

func (s *PriceCategoryService) DeletePriceCategory(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.PriceCategoryModel{}).Error
}
