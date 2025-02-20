package video

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type VideoService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewVideoService(db *gorm.DB, ctx *context.ERPContext) *VideoService {
	return &VideoService{
		db:  db,
		ctx: ctx,
	}
}

func (s *VideoService) CreateVideo(data *models.VideoModel) error {
	return s.db.Create(data).Error
}

func (s *VideoService) UpdateVideo(id string, data *models.VideoModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

func (s *VideoService) DeleteVideo(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.VideoModel{}).Error
}

func (s *VideoService) GetVideoByID(id string) (*models.VideoModel, error) {
	var article models.VideoModel
	err := s.db.Where("id = ?", id).First(&article).Error
	return &article, err
}

func (s *VideoService) GetVideos(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db
	if search != "" {
		stmt = stmt.Where("videos.description ILIKE ? OR videos.title ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
		)
	}
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	request.URL.Query().Get("page")
	stmt = stmt.Model(&models.VideoModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.VideoModel{})
	page.Page = page.Page + 1
	return page, nil
}
