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

func (cs *ContentManagementService) Migrate() error {
	if cs.ctx.SkipMigration {
		return nil
	}
	return cs.ctx.DB.AutoMigrate(&models.ArticleModel{}, &models.ContentCategoryModel{}, &models.ContentCommentModel{})
}
