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

// NewArticleService returns a new instance of ArticleService.
//
// The service is initialized with a GORM database instance and an ERP context.
// The database instance is utilized for handling CRUD operations related to articles,
// while the ERP context is used for authentication and authorization purposes.

func NewArticleService(db *gorm.DB, ctx *context.ERPContext) *ArticleService {
	return &ArticleService{
		db:  db,
		ctx: ctx,
	}
}

// CreateArticle creates a new article in the database.
//
// The function takes a pointer to an ArticleModel, which contains the data to be
// inserted into the database. It returns an error if the insertion fails, or
// nil if the insertion is successful.
func (s *ArticleService) CreateArticle(data *models.ArticleModel) error {
	return s.db.Create(data).Error
}

// UpdateArticle updates an existing article in the database.
//
// It takes an ID of the article to be updated and a pointer to an ArticleModel
// containing the new data. The function returns an error if the update operation
// fails. If successful, the article is updated with the provided data and the error
// is nil.

func (s *ArticleService) UpdateArticle(id string, data *models.ArticleModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

// DeleteArticle deletes an article from the database by its ID.
//
// It takes an ID as input and returns an error if the deletion operation fails.
// The function uses GORM to delete the article data from the articles table.
// If the deletion is successful, the error is nil. Otherwise, the error contains
// information about what went wrong.

func (s *ArticleService) DeleteArticle(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.ArticleModel{}).Error
}

// GetArticleByID retrieves an article from the database by its ID.
//
// It takes an ID as input and returns an ArticleModel and an error. The
// function uses GORM to retrieve the article data from the articles table.
// If the operation fails, an error is returned.
func (s *ArticleService) GetArticleByID(id string) (*models.ArticleModel, error) {
	var article models.ArticleModel
	err := s.db.Where("id = ?", id).First(&article).Error
	return &article, err
}

// GetArticles retrieves a paginated list of articles from the database.
//
// It accepts an HTTP request and a search query string as parameters. The search query
// is applied to the article's description and title fields. If a company ID is present
// in the request header, the result is filtered by the company ID. The function uses
// pagination to manage the result set and includes any necessary request modifications
// using the utils.FixRequest utility.
//
// The function returns a paginated page of ArticleModel and an error if the operation fails.

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
