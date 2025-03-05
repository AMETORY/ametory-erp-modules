package models

import (
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CrowdFundingDonationModel struct {
	shared.BaseModel
	Campaign      CrowdFundingCampaignModel `gorm:"foreignKey:CampaignID"`
	CampaignID    string                    `gorm:"type:varchar(36);index;not null;constraint:OnDelete:CASCADE" json:"campaign_id,omitempty"`
	Donor         *UserModel                `gorm:"foreignKey:DonorID"`
	DonorID       *string                   `gorm:"type:varchar(36);index;constraint:OnDelete:CASCADE" json:"donor_id,omitempty"`
	Amount        float64                   `gorm:"type:decimal(13,2)" json:"amount,omitempty"`
	PaymentMethod string                    `json:"payment_method,omitempty"`
	Status        string                    `gorm:"type:varchar(20)" json:"status,omitempty"`
	Payment       PaymentModel              `gorm:"foreignKey:PaymentID"`
	PaymentID     *string                   `gorm:"type:varchar(36);index;constraint:OnDelete:CASCADE" json:"payment_id,omitempty"`
}

func (CrowdFundingDonationModel) TableName() string {
	return "crowd_funding_donations"
}

func (model *CrowdFundingDonationModel) BeforeCreate(tx *gorm.DB) (err error) {
	// Add any logic you want to execute before creating a record

	if model.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return nil
}
