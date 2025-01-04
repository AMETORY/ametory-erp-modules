package product

import (
	"net/http"

	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type ProductService struct {
	db *gorm.DB
}
type ProductCategoryService struct {
	db *gorm.DB
}

func NewProductService(db *gorm.DB) *ProductService {
	return &ProductService{db: db}
}
func NewProductCategoryService(db *gorm.DB) *ProductCategoryService {
	return &ProductCategoryService{db: db}
}

func (s *ProductService) CreateProduct(data *ProductModel) error {
	return s.db.Create(data).Error
}

func (s *ProductService) UpdateProduct(id string, data *ProductModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

func (s *ProductService) DeleteProduct(id string) error {
	return s.db.Where("id = ?", id).Delete(&ProductModel{}).Error
}

func (s *ProductService) GetProductByID(id string) (*ProductModel, error) {
	var invoice ProductModel
	err := s.db.Where("id = ?", id).First(&invoice).Error
	return &invoice, err
}

func (s *ProductService) GetProductByCode(code string) (*ProductModel, error) {
	var invoice ProductModel
	err := s.db.Where("code = ?", code).First(&invoice).Error
	return &invoice, err
}

func (s *ProductService) GetProducts(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db
	if search != "" {
		stmt = stmt.Where("products.description LIKE ? OR products.sku LIKE ? OR products.name LIKE ? OR products.barcode LIKE ?",
			"%"+search+"%",
			"%"+search+"%",
			"%"+search+"%",
			"%"+search+"%",
		)
	}
	stmt = stmt.Model(&ProductModel{})
	page := pg.With(stmt).Request(request).Response(&[]ProductModel{})
	return page, nil
}

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
	page := pg.With(stmt).Request(request).Response(&[]ProductCategoryModel{})
	return page, nil
}
