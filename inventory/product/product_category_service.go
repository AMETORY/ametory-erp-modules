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

// NewProductCategoryService creates a new instance of ProductCategoryService.
//
// Args:
//
//	db: the gorm database instance.
//	ctx: the erp context.
//
// Returns:
//
//	A new instance of ProductCategoryService.
func NewProductCategoryService(db *gorm.DB, ctx *context.ERPContext) *ProductCategoryService {
	return &ProductCategoryService{db: db, ctx: ctx}
}

// CreateProductCategory creates a new product category in the database.
//
// It takes a pointer to a ProductCategoryModel as a parameter and returns an error.
// The error will be nil if the product category was created successfully.
func (s *ProductCategoryService) CreateProductCategory(data *models.ProductCategoryModel) error {
	return s.db.Create(data).Error
}

// UpdateProductCategory updates an existing product category.
//
// Args:
//
//	id: the id of the product category to update.
//	data: the updated data of the product category.
//
// Returns:
//
//	an error if any error occurs.
func (s *ProductCategoryService) UpdateProductCategory(id string, data *models.ProductCategoryModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

// DeleteProductCategory deletes a product category by its ID.
//
// It takes a string id as a parameter and returns an error if the deletion fails.
//
// It uses the gorm.DB connection to delete a record from the product_categories table.
func (s *ProductCategoryService) DeleteProductCategory(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.ProductCategoryModel{}).Error
}

// GetProductCategoryByID retrieves a product category by its ID.
//
// It takes a string id as a parameter and returns a pointer to a ProductCategoryModel
// and an error. The function uses GORM to retrieve the product category data from
// the product_categories table. If the operation fails, an error is returned.
func (s *ProductCategoryService) GetProductCategoryByID(id string) (*models.ProductCategoryModel, error) {
	var category models.ProductCategoryModel
	err := s.db.Where("id = ?", id).First(&category).Error
	return &category, err
}

// GetProductCategories retrieves a paginated list of product categories from the database.
//
// It takes an http.Request and a search query string as input. The method uses
// GORM to query the database for product categories, applying the search query to
// the product category name and description fields. If the request contains a
// company ID header, the method also filters the result by the company ID.
// The function utilizes pagination to manage the result set and applies any
// necessary request modifications using the utils.FixRequest utility.
//
// The function returns a paginated page of ProductCategoryModel and an error if
// the operation fails.
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

// GetCategoryByName retrieves a product category by its name.
//
// If the category does not exist, a new one is created with the given name.
//
// It takes a string name as a parameter and returns a pointer to a ProductCategoryModel
// and an error. The function uses GORM to retrieve the product category data from
// the product_categories table. If the operation fails, an error is returned.
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
