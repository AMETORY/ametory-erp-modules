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

// CreateDonation creates a new donation record in the database.
//
// It takes a pointer to a CrowdFundingDonationModel as a parameter and returns
// an error if the creation fails.

func (s *CrowdFundingDonationService) CreateDonation(data *models.CrowdFundingDonationModel) error {
	return s.ctx.DB.Create(data).Error
}

// PaymentDonation performs a transaction to create a new payment record
// in the database and updates the donation record with the new payment
// status and payment method.
//
// It takes the ID of the donation record and a pointer to a PaymentModel
// as parameters and returns an error if the transaction fails.
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

// GetDonationByCampaignID retrieves a paginated list of donations for a given campaign ID.
//
// It takes an HTTP request and a campaign ID string as parameters. The donations
// are filtered by the specified campaign ID. The function uses pagination to
// manage the result set and includes any necessary request modifications using
// the utils.FixRequest utility.
//
// The function returns a paginated page of CrowdFundingDonationModel and an error
// if the operation fails.

func (s *CrowdFundingDonationService) GetDonationByCampaignID(request *http.Request, campaignID string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.ctx.DB.Where("campaign_id = ?", campaignID)
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.CrowdFundingDonationModel{})
	page.Page = page.Page + 1
	return page, nil
}

// GetDonationByUserID retrieves a paginated list of donations for a given user ID.
//
// It takes an HTTP request and a user ID string as parameters. The donations
// are filtered by the specified user ID. The function uses pagination to
// manage the result set and includes any necessary request modifications using
// the utils.FixRequest utility.
//
// The function returns a paginated page of CrowdFundingDonationModel and an error
// if the operation fails.
func (s *CrowdFundingDonationService) GetDonationByUserID(request *http.Request, userID string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.ctx.DB.Where("donor_id = ?", userID)
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.CrowdFundingDonationModel{})
	page.Page = page.Page + 1
	return page, nil
}
