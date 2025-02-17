package banner

import (
	"errors"
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type BannerService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewBannerService(db *gorm.DB, ctx *context.ERPContext) *BannerService {
	return &BannerService{db: db, ctx: ctx}
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.BannerModel{},
	)
}

func (s *BannerService) CreateBanner(data *models.BannerModel) error {
	return s.db.Create(data).Error
}

func (s *BannerService) UpdateBanner(id string, data *models.BannerModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

func (s *BannerService) DeleteBanner(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.BannerModel{}).Error
}

func (s *BannerService) GetBannerByID(id string) (*models.BannerModel, error) {
	var invoice models.BannerModel
	err := s.db.Where("id = ?", id).First(&invoice).Error
	return &invoice, err
}

func (s *BannerService) GetBannerByCode(code string) (*models.BannerModel, error) {
	var invoice models.BannerModel
	err := s.db.Where("code = ?", code).First(&invoice).Error
	return &invoice, err
}

func (s *BannerService) GetBanners(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db
	if search != "" {
		stmt = stmt.Where("banners.description ILIKE ? OR banners.title ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
		)
	}
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	request.URL.Query().Get("page")
	stmt = stmt.Model(&models.BannerModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.BannerModel{})
	page.Page = page.Page + 1
	return page, nil
}

func (s *BannerService) GetBannerByName(title string) (*models.BannerModel, error) {
	var banner models.BannerModel
	err := s.db.Where("title = ?", title).First(&banner).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			banner = models.BannerModel{
				Title: title,
			}
			err := s.db.Create(&banner).Error
			if err != nil {
				return nil, err
			}
			return &banner, nil
		}
		return nil, err
	}
	return &banner, nil
}
