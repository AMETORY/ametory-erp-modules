package product

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
)

func (s *ProductCategoryService) CreateProductCategory(data *ProductCategoryModel) error {
	return s.db.Create(data).Error
}

func (s *ProductCategoryService) UpdateProductCategory(id string, data *ProductCategoryModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

func (s *ProductCategoryService) DeleteProductCategory(id string) error {
	return s.db.Where("id = ?", id).Delete(&ProductCategoryModel{}).Error
}

func (s *ProductCategoryService) GetProductCategoryByID(id string) (*ProductCategoryModel, error) {
	var category ProductCategoryModel
	err := s.db.Where("id = ?", id).First(&category).Error
	return &category, err
}

func (s *ProductCategoryService) GetProductCategories(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db
	if search != "" {
		stmt = stmt.Where("product_categories.name LIKE ? OR product_categories.description LIKE ?",
			"%"+search+"%",
			"%"+search+"%",
		)
	}
	stmt = stmt.Model(&ProductCategoryModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]ProductCategoryModel{})
	page.Page = page.Page + 1
	return page, nil
}
