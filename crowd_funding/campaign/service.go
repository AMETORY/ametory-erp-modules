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

// NewCrowdFundingCampaignService returns a new instance of CrowdFundingCampaignService.
//
// The CrowdFundingCampaignService is used to manage crowd funding campaigns.
//
// It takes as parameters:
//   - db: a pointer to a GORM database instance
//   - ctx: a pointer to an ERPContext, which contains the user's HTTP request
//     context and other relevant information
func NewCrowdFundingCampaignService(db *gorm.DB, ctx *context.ERPContext) *CrowdFundingCampaignService {
	return &CrowdFundingCampaignService{
		db:  db,
		ctx: ctx,
	}
}

// CreateCampaign creates a new crowd funding campaign.
//
// It takes a pointer to a CrowdFundingCampaignModel as a parameter and returns
// an error if the creation fails.
//
// The request context is used to determine the user who is creating the campaign.
func (s *CrowdFundingCampaignService) CreateCampaign(data *models.CrowdFundingCampaignModel) error {
	return s.ctx.DB.Create(data).Error
}

// UpdateCampaign updates a crowd funding campaign.
//
// It takes a pointer to a CrowdFundingCampaignModel as a parameter and returns
// an error if the update fails.
//
// The request context is used to determine the user who is updating the campaign.
func (s *CrowdFundingCampaignService) UpdateCampaign(data *models.CrowdFundingCampaignModel) error {
	return s.ctx.DB.Save(data).Error
}

// DeleteCampaign deletes a crowd funding campaign by its ID.
//
// It takes the campaign ID as a parameter and returns an error if the deletion fails.
// The request context is used to perform the deletion operation.

func (s *CrowdFundingCampaignService) DeleteCampaign(id string) error {
	return s.ctx.DB.Delete(&models.CrowdFundingCampaignModel{}, "id = ?", id).Error
}
