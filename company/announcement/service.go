package announcement

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type AnnoucementService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewAnnoucementService(db *gorm.DB, ctx *context.ERPContext) *AnnoucementService {
	return &AnnoucementService{db: db, ctx: ctx}
}

func (s *AnnoucementService) CreateAnnoucement(data *models.AnnoucementModel) error {
	return s.db.Create(data).Error
}

func (s *AnnoucementService) UpdateAnnoucement(id string, data *models.AnnoucementModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

func (s *AnnoucementService) DeleteAnnoucement(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.AnnoucementModel{}).Error
}

func (s *AnnoucementService) GetAnnoucementByID(id string) (*models.AnnoucementModel, error) {
	var branch models.AnnoucementModel
	if err := s.db.Where("id = ?", id).First(&branch).Error; err != nil {
		return nil, err
	}
	return &branch, nil
}

func (s *AnnoucementService) FindAllAnnoucementes(request *http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.Model(&models.AnnoucementModel{})
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.AnnoucementModel{})
	page.Page = page.Page + 1
	return page, nil
}
