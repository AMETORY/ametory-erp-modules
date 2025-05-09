package product

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type ProductCategoryService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewProductCategoryService(db *gorm.DB, ctx *context.ERPContext) *ProductCategoryService {
	return &ProductCategoryService{db: db, ctx: ctx}
}

func (s *ProductCategoryService) CreateProductCategory(data *models.ProductCategoryModel) error {
	return s.db.Create(data).Error
}

func (s *ProductCategoryService) UpdateProductCategory(id string, data *models.ProductCategoryModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

func (s *ProductCategoryService) DeleteProductCategory(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.ProductCategoryModel{}).Error
}

func (s *ProductCategoryService) GetProductCategoryByID(id string) (*models.ProductCategoryModel, error) {
	var category models.ProductCategoryModel
	err := s.db.Where("id = ?", id).First(&category).Error
	return &category, err
}

func (s *ProductCategoryService) GetProductCategories(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db
	if search != "" {
		stmt = stmt.Where("product_categories.name ILIKE ? OR product_categories.description ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
		)
	}

	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ? or company_id is null", request.Header.Get("ID-Company"))
	}
	stmt = stmt.Model(&models.ProductCategoryModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.ProductCategoryModel{})
	page.Page = page.Page + 1
	return page, nil
}

func (s *ProductCategoryService) GetCategoryByName(name string) (*models.ProductCategoryModel, error) {
	var category models.ProductCategoryModel
	err := s.db.Where("name = ?", name).First(&category).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	if err == gorm.ErrRecordNotFound {
		category.Name = name
		err = s.CreateProductCategory(&category)
		if err != nil {
			return nil, err
		}
	}
	return &category, nil
}
