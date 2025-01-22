package product

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"gorm.io/gorm"
)

type TagService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewTagService(db *gorm.DB, ctx *context.ERPContext) *TagService {
	return &TagService{db: db, ctx: ctx}
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
	if err := s.db.Model(&models.TagModel{}).Where("id = ?", id).Save(tag).Error; err != nil {
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

func (s *TagService) ListTags() ([]models.TagModel, error) {
	var tags []models.TagModel
	if err := s.db.Find(&tags).Error; err != nil {
		return nil, err
	}
	return tags, nil
}
