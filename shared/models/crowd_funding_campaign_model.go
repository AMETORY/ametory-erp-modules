package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CrowdFundingCampaignModel struct {
	shared.BaseModel
	Title         string                             `gorm:"type:varchar(255);not null" json:"title,omitempty"`
	Description   string                             `gorm:"type:text" json:"description,omitempty"`
	TargetAmount  float64                            `gorm:"not null" json:"target_amount,omitempty"`
	CurrentAmount float64                            `gorm:"not null" json:"current_amount,omitempty"`
	Deadline      *time.Time                         `json:"deadline,omitempty"`
	Images        []FileModel                        `gorm:"-" json:"images,omitempty"`
	CreatedByID   *string                            `gorm:"type:char(36);index" json:"created_by_id,omitempty"`
	CreatedBy     *UserModel                         `gorm:"foreignKey:CreatedByID;constraint:OnDelete:CASCADE;" json:"created_by,omitempty"`
	Status        string                             `gorm:"type:varchar(20);default:'active'" json:"status,omitempty"`
	Donors        []CrowdFundingDonationModel        `gorm:"foreignKey:CampaignID;constraint:OnDelete:CASCADE" json:"donors,omitempty"`
	Tag           []TagModel                         `gorm:"many2many:crowd_funding_campaign_tag;" json:"tags,omitempty"`
	CategoryID    *string                            `gorm:"type:char(36);index" json:"category_id,omitempty"`
	Category      *CrowdFundingCampaignCategoryModel `gorm:"foreignKey:CategoryID;constraint:OnDelete:CASCADE;" json:"category,omitempty"`
}

func (CrowdFundingCampaignModel) TableName() string {
	return "crowd_funding_campaigns"
}

func (c *CrowdFundingCampaignModel) BeforeCreate(tx *gorm.DB) (err error) {
	c.ID = uuid.NewString() // generate a random ID
	// Add any logic you need to execute before creating a record
	return nil
}

func (c *CrowdFundingCampaignModel) AfterFind(tx *gorm.DB) (err error) {
	var images []FileModel
	err = tx.Where("ref_id = ? and ref_type = ?", c.ID, "crowd_funding_campaign").Find(&images).Error
	c.Images = images
	return err
}

type CrowdFundingCampaignCategoryModel struct {
	shared.BaseModel
	Name string `gorm:"type:varchar(255);not null;unique" json:"name,omitempty"`
}

func (CrowdFundingCampaignCategoryModel) TableName() string {
	return "crowd_funding_campaign_categories"
}

func (c *CrowdFundingCampaignCategoryModel) BeforeCreate(tx *gorm.DB) (err error) {
	c.ID = uuid.NewString() // generate a random ID
	// Add any logic you need to execute before creating a record
	return nil
}
