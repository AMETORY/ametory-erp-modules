package article

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type ArticleService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewArticleService(db *gorm.DB, ctx *context.ERPContext) *ArticleService {
	return &ArticleService{
		db:  db,
		ctx: ctx,
	}
}

func (s *ArticleService) CreateArticle(data *models.ArticleModel) error {
	return s.db.Create(data).Error
}

func (s *ArticleService) UpdateArticle(id string, data *models.ArticleModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

func (s *ArticleService) DeleteArticle(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.ArticleModel{}).Error
}

func (s *ArticleService) GetArticleByID(id string) (*models.ArticleModel, error) {
	var article models.ArticleModel
	err := s.db.Where("id = ?", id).First(&article).Error
	return &article, err
}

func (s *ArticleService) GetArticles(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db
	if search != "" {
		stmt = stmt.Where("articles.description ILIKE ? OR articles.title ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
		)
	}
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	request.URL.Query().Get("page")
	stmt = stmt.Model(&models.ArticleModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.ArticleModel{})
	page.Page = page.Page + 1
	return page, nil
}
