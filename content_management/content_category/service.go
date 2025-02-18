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

func NewContentCategoryService(db *gorm.DB, ctx *context.ERPContext) *ContentCategoryService {
	return &ContentCategoryService{
		db:  db,
		ctx: ctx,
	}
}
func (as *ContentCategoryService) Migrate() error {
	return as.db.AutoMigrate(&models.ContentCategoryModel{}, &models.ContentCategoryModel{}, &models.ContentCommentModel{})
}

func (s *ContentCategoryService) CreateContentCategory(data *models.ContentCategoryModel) error {
	return s.db.Create(data).Error
}

func (s *ContentCategoryService) UpdateContentCategory(id string, data *models.ContentCategoryModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

func (s *ContentCategoryService) DeleteContentCategory(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.ContentCategoryModel{}).Error
}

func (s *ContentCategoryService) GetContentCategoryByID(id string) (*models.ContentCategoryModel, error) {
	var article models.ContentCategoryModel
	err := s.db.Where("id = ?", id).First(&article).Error
	return &article, err
}

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
