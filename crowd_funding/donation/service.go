package donation

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"gorm.io/gorm"
)

type CrowdFundingDonationService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewCrowdFundingDonationService(db *gorm.DB, ctx *context.ERPContext) *CrowdFundingDonationService {
	return &CrowdFundingDonationService{
		db:  db,
		ctx: ctx,
	}
}
