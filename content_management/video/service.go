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

// NewVideoService returns a new instance of VideoService.
//
// The returned instance is properly initialized with the given database
// connection and context.
func NewVideoService(db *gorm.DB, ctx *context.ERPContext) *VideoService {
	return &VideoService{
		db:  db,
		ctx: ctx,
	}
}

// CreateVideo creates a new video in the database.
//
// The given data is inserted into the videos table. The ID field in the given
// data is ignored and set automatically by GORM.
//
// The function returns an error if the insertion fails.
func (s *VideoService) CreateVideo(data *models.VideoModel) error {
	return s.db.Create(data).Error
}

// UpdateVideo updates an existing video in the database.
//
// The given id is used to find the video in the database. The given data is
// used to update the video. The ID field in the given data is ignored and
// cannot be updated.
//
// The function returns an error if the update fails.
func (s *VideoService) UpdateVideo(id string, data *models.VideoModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

// DeleteVideo deletes an existing video from the database.
//
// The given id is used to find the video in the database. The function returns
// an error if the deletion fails.
func (s *VideoService) DeleteVideo(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.VideoModel{}).Error
}

// GetVideoByID retrieves a video from the database by its ID.
//
// It takes an ID as input and returns a pointer to a VideoModel and an error.
// The function uses GORM to retrieve the video data from the videos table. If
// the operation fails, an error is returned.
func (s *VideoService) GetVideoByID(id string) (*models.VideoModel, error) {
	var article models.VideoModel
	err := s.db.Where("id = ?", id).First(&article).Error
	return &article, err
}

// GetVideos retrieves a paginated list of videos from the database.
//
// It takes an HTTP request and a search query string as input. The search
// query is applied to the video's description and title fields. If a company
// ID is present in the request header, the result is filtered by the company
// ID. The function utilizes pagination to manage the result set and applies
// any necessary request modifications using the utils.FixRequest utility.
//
// The function returns a paginated page of VideoModel and an error if the
// operation fails.
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
