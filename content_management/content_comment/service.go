package content_comment

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type ContentCommentService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

// NewContentCommentService creates a new instance of ContentCommentService.
//
// It initializes the service with a GORM database instance and an ERP context.
// The database instance is used for executing database operations related to
// content comments, while the ERP context is used for handling authentication
// and authorization processes.

func NewContentCommentService(db *gorm.DB, ctx *context.ERPContext) *ContentCommentService {
	return &ContentCommentService{
		db:  db,
		ctx: ctx,
	}
}

// CreateContentComment creates a new content comment.
//
// The function takes a pointer to a ContentCommentModel instance, which contains
// the data to be inserted into the database. It returns an error if the insertion
// fails.
func (s *ContentCommentService) CreateContentComment(data *models.ContentCommentModel) error {
	return s.db.Create(data).Error
}

// UpdateContentComment updates an existing content comment.
//
// It takes a string id and a pointer to a ContentCommentModel as parameters and
// returns an error. It uses the gorm.DB connection to update a record in the
// content_comments table.
func (s *ContentCommentService) UpdateContentComment(id string, data *models.ContentCommentModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

// DeleteContentComment deletes a content comment.
//
// It takes a string id as a parameter and deletes a record in the
// content_comments table with the given id. It returns an error if the deletion
// fails.
func (s *ContentCommentService) DeleteContentComment(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.ContentCommentModel{}).Error
}

// GetContentCommentByID retrieves a content comment by its ID.
//
// It takes a string id as a parameter and returns a pointer to a ContentCommentModel
// and an error. The function uses GORM to retrieve the content comment data from
// the content_comments table. If the operation fails, an error is returned.

func (s *ContentCommentService) GetContentCommentByID(id string) (*models.ContentCommentModel, error) {
	var article models.ContentCommentModel
	err := s.db.Where("id = ?", id).First(&article).Error
	return &article, err
}

// GetContentComments retrieves a paginated list of content comments from the database.
//
// It takes an HTTP request and a search query string as parameters. The search query
// is applied to the content comment's comment field. If a company ID is present in the
// request header, the result is filtered by the company ID. The function uses
// pagination to manage the result set and includes any necessary request modifications
// using the utils.FixRequest utility.
//
// The function returns a paginated page of ContentCommentModel and an error if the
// operation fails.
func (s *ContentCommentService) GetContentComments(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db
	if search != "" {
		stmt = stmt.Where("comment ILIKE ?",
			"%"+search+"%",
		)
	}
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	request.URL.Query().Get("page")
	stmt = stmt.Model(&models.ContentCommentModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.ContentCommentModel{})
	page.Page = page.Page + 1
	return page, nil
}
