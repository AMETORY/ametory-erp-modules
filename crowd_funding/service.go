package crowd_funding

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/crowd_funding/campaign"
	"github.com/AMETORY/ametory-erp-modules/crowd_funding/donation"
)

type CrowdFundingService struct {
	ctx                         *context.ERPContext
	CrowdFundingCampaignService *campaign.CrowdFundingCampaignService
	CrowdFundingDonationService *donation.CrowdFundingDonationService
}

func NewCrowdFundingService(ctx *context.ERPContext) *CrowdFundingService {
	crowdFundingSrv := CrowdFundingService{
		CrowdFundingCampaignService: campaign.NewCrowdFundingCampaignService(ctx.DB, ctx),
		CrowdFundingDonationService: donation.NewCrowdFundingDonationService(ctx.DB, ctx),
		// ContentComment:  content_comment.NewContentCommentService(ctx.DB, ctx),
		ctx: ctx,
	}
	err := crowdFundingSrv.Migrate()
	if err != nil {
	}

	return &crowdFundingSrv
}

func (cs *CrowdFundingService) Migrate() error {
	if cs.ctx.SkipMigration {
		return nil
	}
	return cs.ctx.DB.AutoMigrate()
}
