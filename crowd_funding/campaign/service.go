package campaign

import (
	"github.com/AMETORY/ametory-erp-modules/context"
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
