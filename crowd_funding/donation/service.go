package donation

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
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

func (s *CrowdFundingDonationService) CreateDonation(data *models.CrowdFundingDonationModel) error {
	return s.ctx.DB.Create(data).Error
}

func (s *CrowdFundingDonationService) PaymentDonation(donationID string, data *models.PaymentModel) error {
	return s.ctx.DB.Transaction(func(tx *gorm.DB) error {
		err := s.ctx.DB.Create(data).Error
		if err != nil {
			return err
		}

		var donation models.CrowdFundingDonationModel
		err = tx.Where("id = ?", donationID).First(&donation).Error
		if err != nil {
			return err
		}
		donation.Status = data.Status
		donation.PaymentMethod = data.PaymentMethod
		donation.PaymentID = &data.ID
		err = tx.Save(&donation).Error
		if err != nil {
			return err
		}
		return nil
	})
}

func (s *CrowdFundingDonationService) GetDonationByCampaignID(request *http.Request, campaignID string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.ctx.DB.Where("campaign_id = ?", campaignID)
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.CrowdFundingDonationModel{})
	page.Page = page.Page + 1
	return page, nil
}

func (s *CrowdFundingDonationService) GetDonationByUserID(request *http.Request, userID string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.ctx.DB.Where("donor_id = ?", userID)
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.CrowdFundingDonationModel{})
	page.Page = page.Page + 1
	return page, nil
}
