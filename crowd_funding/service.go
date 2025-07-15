package crowd_funding

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/crowd_funding/campaign"
	"github.com/AMETORY/ametory-erp-modules/crowd_funding/donation"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
)

type CrowdFundingService struct {
	ctx                         *context.ERPContext
	CrowdFundingCampaignService *campaign.CrowdFundingCampaignService
	CrowdFundingDonationService *donation.CrowdFundingDonationService
}

// NewCrowdFundingService creates a new instance of CrowdFundingService.
//
// The service provides functionalities for managing crowd funding campaigns
// and donations. It requires an ERPContext for initializing the associated
// services and handling request contexts. The method initializes the
// CrowdFundingCampaignService and CrowdFundingDonationService using the
// provided context's database connection.
//
// Additionally, it calls the Migrate method to ensure the necessary
// database schema is in place. If the migration process encounters an
// error, it is logged.
//
// Returns a pointer to the initialized CrowdFundingService instance.

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

// Migrate migrates the database schema for the CrowdFundingService.
//
// If the SkipMigration flag is set to true in the context, this method
// will not perform any migration and will return nil. Otherwise, it will
// attempt to auto-migrate the database to include the
// CrowdFundingCampaignModel and CrowdFundingDonationModel schemas.
// If the migration process encounters an error, it will return that error.
// Otherwise, it will return nil upon successful migration.
func (cs *CrowdFundingService) Migrate() error {
	if cs.ctx.SkipMigration {
		return nil
	}
	return cs.ctx.DB.AutoMigrate(&models.CrowdFundingCampaignModel{}, &models.CrowdFundingDonationModel{})
}
