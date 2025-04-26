package tag

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type TagService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewTagService(ctx *context.ERPContext) *TagService {
	if !ctx.SkipMigration {
		ctx.DB.AutoMigrate(&models.TagModel{})
	}
	return &TagService{db: ctx.DB, ctx: ctx}
}

func (s *TagService) CreateTag(tag *models.TagModel) error {
	if err := s.db.Create(tag).Error; err != nil {
		return err
	}
	return nil
}

func (s *TagService) GetTagByID(id string) (*models.TagModel, error) {
	var tag models.TagModel
	if err := s.db.First(&tag, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &tag, nil
}

func (s *TagService) GetTagByName(name string) (*models.TagModel, error) {
	var tag models.TagModel
	if err := s.db.Where("name ILIKE ?", "%"+name+"%").First(&tag).Error; err != nil {
		return nil, err
	}
	return &tag, nil
}

func (s *TagService) UpdateTag(id string, tag *models.TagModel) error {
	if err := s.db.Model(&models.TagModel{}).Where("id = ?", id).Updates(tag).Error; err != nil {
		return err
	}
	return nil
}

func (s *TagService) DeleteTag(id string) error {
	if err := s.db.Delete(&models.TagModel{}, id).Error; err != nil {
		return err
	}
	return nil
}

func (s *TagService) ListTags(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db
	stmt = stmt.Model(&models.TagModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.TagModel{})
	page.Page = page.Page + 1
	return page, nil
}
