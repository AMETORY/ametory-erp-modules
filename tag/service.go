package tag

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

// TagService provides methods for managing tags within the application.
//
// NewTagService creates a new instance of TagService.
//
// CreateTag creates a new tag in the database.
//
// GetTagByID retrieves a tag by its ID.
//
// GetTagByName retrieves a tag by its name.
//
// UpdateTag updates an existing tag in the database.
//
// DeleteTag deletes a tag from the database by its ID.
//
// ListTags retrieves a paginated list of tags from the database.
//
// The service is initialized with a GORM database instance and an ERP context.
type TagService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

// NewTagService creates a new instance of TagService.
//
// It requires a GORM database instance and an ERP context for initialization.
// The TagService provides methods for managing tags within the application.
// The database instance is used for executing CRUD operations, while the
// ERP context is utilized for handling authentication and other context-related
// functionality.
//
// Returns a pointer to a newly created TagService.
func NewTagService(ctx *context.ERPContext) *TagService {
	if !ctx.SkipMigration {
		ctx.DB.AutoMigrate(&models.TagModel{})
	}
	return &TagService{db: ctx.DB, ctx: ctx}
}

// CreateTag creates a new tag in the database.
//
// It takes a pointer to a TagModel as input and creates a new record in the tags table.
// If the creation fails, the method returns an error. Otherwise, it returns nil.
func (s *TagService) CreateTag(tag *models.TagModel) error {
	if err := s.db.Create(tag).Error; err != nil {
		return err
	}
	return nil
}

// GetTagByID retrieves a tag by its ID.
//
// It takes a string id as a parameter and returns a pointer to a TagModel
// and an error. The function uses GORM to retrieve the tag data from
// the tags table. If the operation fails, an error is returned.
func (s *TagService) GetTagByID(id string) (*models.TagModel, error) {
	var tag models.TagModel
	if err := s.db.First(&tag, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &tag, nil
}

// GetTagByName retrieves a tag by its name.
//
// It takes a string name as a parameter and returns a pointer to a TagModel
// and an error. The function uses GORM to retrieve the tag data from
// the tags table. If the operation fails, an error is returned.
func (s *TagService) GetTagByName(name string) (*models.TagModel, error) {
	var tag models.TagModel
	if err := s.db.Where("name ILIKE ?", "%"+name+"%").First(&tag).Error; err != nil {
		return nil, err
	}
	return &tag, nil
}

// UpdateTag updates an existing tag in the database.
//
// It takes a string id and a pointer to a TagModel as parameters and returns an error.
// The function uses GORM to update the tag data in the database where the tag ID matches.
// If the update is successful, the error is nil. Otherwise, the error contains information about what went wrong.
func (s *TagService) UpdateTag(id string, tag *models.TagModel) error {
	if err := s.db.Model(&models.TagModel{}).Where("id = ?", id).Updates(tag).Error; err != nil {
		return err
	}
	return nil
}

// DeleteTag deletes a tag from the database by its ID.
//
// It takes a string id as a parameter and attempts to delete the corresponding
// TagModel record. The function returns an error if the deletion fails,
// otherwise it returns nil indicating the operation was successful.
func (s *TagService) DeleteTag(id string) error {
	if err := s.db.Unscoped().Delete(&models.TagModel{}, "id = ?", id).Error; err != nil {
		return err
	}
	return nil
}

// ListTags retrieves a paginated list of tags from the database.
//
// It takes an http.Request and an optional search query string as input. The
// method uses GORM to query the database for tags, applying the search query to
// the tag name field. The function utilizes pagination to manage the result set
// and applies any necessary request modifications using the utils.FixRequest
// utility.
//
// The function returns a paginated page of TagModel and an error if the
// operation fails.
func (s *TagService) ListTags(request http.Request, search string) (paginate.Page, error) {

	pg := paginate.New()

	stmt := s.db
	if request.Header.Get("ID-Company") != "" {
		companyID := request.Header.Get("ID-Company")
		stmt = stmt.Where("company_id = ?", companyID)
	}
	stmt = stmt.Model(&models.TagModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.TagModel{})
	page.Page = page.Page + 1
	return page, nil
}
