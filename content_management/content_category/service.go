package content_category

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type ContentCategoryService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

// NewContentCategoryService creates a new instance of ContentCategoryService.
// It takes a gorm.DB connection and an ERPContext as parameters and returns
// a pointer to a ContentCategoryService.

func NewContentCategoryService(db *gorm.DB, ctx *context.ERPContext) *ContentCategoryService {
	return &ContentCategoryService{
		db:  db,
		ctx: ctx,
	}
}

// CreateContentCategory creates a new content category.
//
// It takes a pointer to a ContentCategoryModel as parameter and returns an error.
//
// It uses the gorm.DB connection to create a new record in the content_categories table.
func (s *ContentCategoryService) CreateContentCategory(data *models.ContentCategoryModel) error {
	return s.db.Create(data).Error
}

// UpdateContentCategory updates an existing content category.
//
// It takes a string id and a pointer to a ContentCategoryModel as parameters and returns an error.
//
// It uses the gorm.DB connection to update a record in the content_categories table.
func (s *ContentCategoryService) UpdateContentCategory(id string, data *models.ContentCategoryModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

// DeleteContentCategory deletes a content category by its ID.
//
// It takes a string id as a parameter and returns an error if the deletion fails.
//
// It uses the gorm.DB connection to delete a record from the content_categories table.

func (s *ContentCategoryService) DeleteContentCategory(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.ContentCategoryModel{}).Error
}

// GetContentCategoryByID retrieves a content category by its ID.
//
// It takes a string id as a parameter and returns a pointer to a ContentCategoryModel
// and an error. The function uses GORM to retrieve the content category data from
// the content_categories table. If the operation fails, an error is returned.
func (s *ContentCategoryService) GetContentCategoryByID(id string) (*models.ContentCategoryModel, error) {
	var article models.ContentCategoryModel
	err := s.db.Where("id = ?", id).First(&article).Error
	return &article, err
}

// GetContentCategorys retrieves a list of content categories with pagination.
//
// It takes an http.Request and a string search as parameters and returns a paginate.Page
// and an error.
//
// The function uses GORM to retrieve the content category data from the content_categories table
// and applies pagination and filtering (based on the search query) to the result set.
// If the operation fails, an error is returned.
func (s *ContentCategoryService) GetContentCategorys(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db
	if search != "" {
		stmt = stmt.Where("name ILIKE ?",
			"%"+search+"%",
		)
	}
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	request.URL.Query().Get("page")
	stmt = stmt.Model(&models.ContentCategoryModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.ContentCategoryModel{})
	page.Page = page.Page + 1
	return page, nil
}
