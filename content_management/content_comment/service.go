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

func NewContentCommentService(db *gorm.DB, ctx *context.ERPContext) *ContentCommentService {
	return &ContentCommentService{
		db:  db,
		ctx: ctx,
	}
}
func (as *ContentCommentService) Migrate() error {
	return as.db.AutoMigrate(&models.ContentCommentModel{}, &models.ContentCommentModel{}, &models.ContentCommentModel{})
}

func (s *ContentCommentService) CreateContentComment(data *models.ContentCommentModel) error {
	return s.db.Create(data).Error
}

func (s *ContentCommentService) UpdateContentComment(id string, data *models.ContentCommentModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

func (s *ContentCommentService) DeleteContentComment(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.ContentCommentModel{}).Error
}

func (s *ContentCommentService) GetContentCommentByID(id string) (*models.ContentCommentModel, error) {
	var article models.ContentCommentModel
	err := s.db.Where("id = ?", id).First(&article).Error
	return &article, err
}

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
