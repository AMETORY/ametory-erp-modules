package campaign

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"gorm.io/gorm"
)

type CrowdFundingCampaignService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewCrowdFundingCampaignService(db *gorm.DB, ctx *context.ERPContext) *CrowdFundingCampaignService {
	return &CrowdFundingCampaignService{
		db:  db,
		ctx: ctx,
	}
}

func (s *CrowdFundingCampaignService) CreateCampaign(data *models.CrowdFundingCampaignModel) error {
	return s.ctx.DB.Create(data).Error
}

func (s *CrowdFundingCampaignService) UpdateCampaign(data *models.CrowdFundingCampaignModel) error {
	return s.ctx.DB.Save(data).Error
}

func (s *CrowdFundingCampaignService) DeleteCampaign(id string) error {
	return s.ctx.DB.Delete(&models.CrowdFundingCampaignModel{}, "id = ?", id).Error
}
