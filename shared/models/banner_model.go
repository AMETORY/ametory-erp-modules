package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BannerModel struct {
	shared.BaseModel
	CompanyID   string        `gorm:"type:char(36);index" json:"company_id,omitempty"`
	Company     CompanyModel  `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE" json:"company,omitempty"`
	ProductID   *string       `gorm:"type:char(36);index" json:"product_id,omitempty"`
	Product     *ProductModel `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE" json:"product,omitempty"`
	VariantID   *string       `gorm:"type:char(36);index" json:"variant_id,omitempty"`
	Variant     *VariantModel `gorm:"foreignKey:VariantID;constraint:OnDelete:CASCADE" json:"variant,omitempty"`
	Title       string        `gorm:"type:varchar(255);not null" json:"title,omitempty"`
	Description string        `gorm:"type:text" json:"description,omitempty"`
	Images      []FileModel   `gorm:"-" json:"image,omitempty"`
	URL         string        `gorm:"type:varchar(255);not null" json:"url,omitempty"`
	StartDate   *time.Time    `json:"start_date,omitempty"`
	EndDate     *time.Time    `json:"end_date,omitempty"`
}

func (BannerModel) TableName() string {
	return "banners"
}

func (b *BannerModel) BeforeCreate(tx *gorm.DB) (err error) {
	b.ID = uuid.New().String()
	return
}

func (b *BannerModel) AfterFind(tx *gorm.DB) (err error) {
	var images []FileModel
	err = tx.Where("ref_id = ? and ref_type = ?", b.ID, "banner").Find(&images).Error
	b.Images = images
	return err
}
