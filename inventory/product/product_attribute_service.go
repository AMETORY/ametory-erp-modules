package product

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type ProductAttributeService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

// NewProductAttributeService creates a new instance of ProductAttributeService.
//
// It requires a gorm.DB instance for database operations and an ERPContext
// for accessing application context and settings. The function returns a pointer
// to a ProductAttributeService, which provides methods for managing product attributes.
func NewProductAttributeService(db *gorm.DB, ctx *context.ERPContext) *ProductAttributeService {
	return &ProductAttributeService{db: db, ctx: ctx}
}

// CreateProductAttribute creates a new product attribute.
//
// The function takes a pointer to a ProductAttributeModel, which contains the data
// for the new product attribute. The function returns an error if the creation
// of the product attribute fails.
func (s *ProductAttributeService) CreateProductAttribute(data *models.ProductAttributeModel) error {
	return s.db.Create(data).Error
}

// UpdateProductAttribute updates an existing product attribute.
//
// The function takes the ID of the product attribute to be updated and a pointer
// to a ProductAttributeModel, which contains the data for the update. The
// function returns an error if the update operation fails.
func (s *ProductAttributeService) UpdateProductAttribute(id string, data *models.ProductAttributeModel) error {
	return s.db.Model(&models.ProductAttributeModel{}).Where("id = ?", id).Updates(data).Error
}

// DeleteProductAttribute deletes a product attribute by its ID.
//
// The function takes the ID of the product attribute to be deleted as a string.
// It returns an error if the deletion operation fails.
func (s *ProductAttributeService) DeleteProductAttribute(id string) error {
	return s.db.Delete(&models.ProductAttributeModel{}, "id = ?", id).Error
}

// GetProductAttributeByID retrieves a product attribute by its ID.
//
// The function takes the ID of the product attribute to be retrieved as a string.
// It returns a pointer to a ProductAttributeModel if the product attribute is found,
// otherwise an error is returned if the retrieval fails.
func (s *ProductAttributeService) GetProductAttributeByID(id string) (*models.ProductAttributeModel, error) {
	var attribute models.ProductAttributeModel
	err := s.db.First(&attribute, "id = ?", id).Error
	return &attribute, err
}

// GetProductAttributes retrieves a paginated list of product attributes from the database.
//
// It takes an http.Request and a search query string as input. The method uses
// GORM to query the database for product attributes, applying the search query to
// the product attribute name field. The result is ordered by the product attribute priority.
// The function utilizes pagination to manage the result set and applies any
// necessary request modifications using the utils.FixRequest utility.
//
// The function returns a paginated page of ProductAttributeModel and an error if
// the operation fails.
func (s *ProductAttributeService) GetProductAttributes(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db
	if search != "" {
		stmt = stmt.Where("product_attributes.name ILIKE ? ",
			"%"+search+"%",
		)
	}
	stmt = stmt.Model(&models.ProductAttributeModel{}).Order("priority ASC")
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.ProductAttributeModel{})
	page.Page = page.Page + 1
	return page, nil
}
