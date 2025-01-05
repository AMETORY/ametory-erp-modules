package brand

import (
	"gorm.io/gorm"

	"github.com/AMETORY/ametory-erp-modules/company"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/google/uuid"
)

// BrandModel adalah model database untuk brand
type BrandModel struct {
	utils.BaseModel
	Name        string `gorm:"not null"`
	Description string
	Logo        string
	CompanyID   string               `json:"company_id"`
	Company     company.CompanyModel `gorm:"foreignKey:CompanyID"`
}

func (BrandModel) TableName() string {
	return "brands"
}

func (bm *BrandModel) BeforeCreate(tx *gorm.DB) (err error) {
	bm.ID = uuid.NewString()
	return
}
