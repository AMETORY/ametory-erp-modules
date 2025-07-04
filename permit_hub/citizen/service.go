package citizen

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
)

type CitizenService struct {
	ctx *context.ERPContext
}

func NewCitizenService(ctx *context.ERPContext) *CitizenService {
	return &CitizenService{
		ctx: ctx,
	}

}

func (s *CitizenService) List() ([]models.Citizen, error) {
	var citizens []models.Citizen
	if err := s.ctx.DB.Find(&citizens).Error; err != nil {
		return nil, err
	}
	return citizens, nil
}

func (s *CitizenService) Create(citizen *models.Citizen) error {
	if err := s.ctx.DB.Create(citizen).Error; err != nil {
		return err
	}
	return nil
}

func (s *CitizenService) Read(nik string) (*models.Citizen, error) {
	var citizen models.Citizen
	if err := s.ctx.DB.Where("nik = ?", nik).First(&citizen).Error; err != nil {
		return nil, err
	}
	return &citizen, nil
}

func (s *CitizenService) Update(citizen *models.Citizen) error {
	return s.ctx.DB.Save(citizen).Error
}

func (s *CitizenService) Delete(id string) error {
	return s.ctx.DB.Where("id = ?", id).Delete(&models.Citizen{}).Error
}
