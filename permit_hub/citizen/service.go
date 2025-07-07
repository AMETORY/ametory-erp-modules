package citizen

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
)

type CitizenService struct {
	ctx *context.ERPContext
}

func NewCitizenService(ctx *context.ERPContext) *CitizenService {
	return &CitizenService{
		ctx: ctx,
	}

}

func (s *CitizenService) GetAllCitizens(request *http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.ctx.DB.Model(&models.Citizen{})
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.Citizen{})
	page.Page = page.Page + 1
	return page, nil
}

func (s *CitizenService) CreateCitizen(citizen *models.Citizen) error {
	if err := s.ctx.DB.Create(citizen).Error; err != nil {
		return err
	}
	return nil
}

func (s *CitizenService) GetCitizenByID(id string) (*models.Citizen, error) {
	var citizen models.Citizen
	if err := s.ctx.DB.Where("id = ?", id).First(&citizen).Error; err != nil {
		return nil, err
	}
	return &citizen, nil
}
func (s *CitizenService) GetCitizenByNIK(nik string) (*models.Citizen, error) {
	var citizen models.Citizen
	if err := s.ctx.DB.Where("nik = ?", nik).First(&citizen).Error; err != nil {
		return nil, err
	}
	return &citizen, nil
}

func (s *CitizenService) UpdateCitizen(id string, citizen *models.Citizen) error {
	return s.ctx.DB.Model(&models.Citizen{}).Where("id = ?", id).Updates(citizen).Error
}

func (s *CitizenService) DeleteCitizen(id string) error {
	return s.ctx.DB.Where("id = ?", id).Delete(&models.Citizen{}).Error
}
