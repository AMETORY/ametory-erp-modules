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

// NewPriceCategoryService creates a new instance of PriceCategoryService with the given database connection and context.
func NewPriceCategoryService(db *gorm.DB, ctx *context.ERPContext) *PriceCategoryService {
	return &PriceCategoryService{db: db, ctx: ctx}
}

// GetPriceCategories retrieves a paginated list of price categories from the database.
//
// It takes an http.Request and a search query string as input. The method uses
// GORM to query the database for price categories, applying the search query to
// the price category name and description fields. If the request contains a
// company ID header, the method also filters the result by the company ID.
// The function utilizes pagination to manage the result set and applies any
// necessary request modifications using the utils.FixRequest utility.
//
// The function returns a paginated page of PriceCategoryModel and an error if
// the operation fails.
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

// GetPriceCategoryByID retrieves a price category from the database by its ID.
//
// It returns a pointer to a PriceCategoryModel and an error if the operation
// fails. If the price category is found, it is returned along with a nil error.
// If not found, or in case of a query error, the function returns a non-nil
// error.
func (s *PriceCategoryService) GetPriceCategoryByID(id string) (*models.PriceCategoryModel, error) {
	var category models.PriceCategoryModel
	err := s.db.Model(&models.PriceCategoryModel{}).Where("id = ?", id).First(&category).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

// CreatePriceCategory creates a new price category in the database.
//
// It takes a pointer to a PriceCategoryModel as parameter and returns an error.
// The error will be nil if the price category was created successfully.
func (s *PriceCategoryService) CreatePriceCategory(data *models.PriceCategoryModel) error {
	return s.db.Create(data).Error
}

// UpdatePriceCategory updates an existing price category.
//
// It takes a string id and a pointer to a PriceCategoryModel as parameters and returns an error.
// The error will be nil if the price category was updated successfully.
func (s *PriceCategoryService) UpdatePriceCategory(id string, data *models.PriceCategoryModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

// DeletePriceCategory deletes a price category from the database by its ID.
//
// It takes a string id as a parameter and attempts to delete the corresponding
// PriceCategoryModel record. The function returns an error if the deletion fails,
// otherwise it returns nil indicating the operation was successful.
func (s *PriceCategoryService) DeletePriceCategory(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.PriceCategoryModel{}).Error
}
