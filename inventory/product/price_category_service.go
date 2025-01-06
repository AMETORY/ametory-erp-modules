package product

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
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

	stmt = stmt.Model(&PriceCategoryModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]PriceCategoryModel{})

	return page, nil
}

func (s *PriceCategoryService) GetPriceCategoryByID(id string) (*PriceCategoryModel, error) {
	var category PriceCategoryModel
	err := s.db.Model(&PriceCategoryModel{}).Where("id = ?", id).First(&category).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}
