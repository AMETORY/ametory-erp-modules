package content_management

import (
	"github.com/AMETORY/ametory-erp-modules/content_management/article"
	"github.com/AMETORY/ametory-erp-modules/content_management/content_category"
	"github.com/AMETORY/ametory-erp-modules/content_management/content_comment"
	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
)

type ContentManagementService struct {
	ctx             *context.ERPContext
	ArticleService  *article.ArticleService
	ContentCategory *content_category.ContentCategoryService
	ContentComment  *content_comment.ContentCommentService
}

// NewContentManagementService creates a new instance of ContentManagementService.
//
// It takes an ERPContext as parameter and returns a pointer to a ContentManagementService.
//
// It uses the ERPContext to initialize the ArticleService, ContentCategoryService, and ContentCommentService.
// Additionally, it calls the Migrate method of the ContentManagementService to create the necessary database schema.
func NewContentManagementService(ctx *context.ERPContext) *ContentManagementService {
	contentManagementSrv := ContentManagementService{
		ArticleService:  article.NewArticleService(ctx.DB, ctx),
		ContentCategory: content_category.NewContentCategoryService(ctx.DB, ctx),
		ContentComment:  content_comment.NewContentCommentService(ctx.DB, ctx),
		ctx:             ctx,
	}
	err := contentManagementSrv.Migrate()
	if err != nil {
	}

	return &contentManagementSrv
}

// Migrate migrates the database schema for the ContentManagementService.
//
// If the SkipMigration flag is set to true in the context, this method
// will not perform any migration and will return nil. Otherwise, it will
// attempt to auto-migrate the database to include the ArticleModel,
// ContentCategoryModel, and ContentCommentModel schemas.
// If the migration process encounters an error, it will return that error.
// Otherwise, it will return nil upon successful migration.

func (cs *ContentManagementService) Migrate() error {
	if cs.ctx.SkipMigration {
		return nil
	}
	return cs.ctx.DB.AutoMigrate(&models.ArticleModel{}, &models.ContentCategoryModel{}, &models.ContentCommentModel{})
}
