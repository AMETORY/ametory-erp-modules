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
	Name               string                `gorm:"not null" json:"name,omitempty"`
	Description        string                `json:"description,omitempty"`
	Logo               string                `json:"logo,omitempty"`
	Website            string                `json:"website,omitempty"`
	Email              string                `json:"email,omitempty"`
	Phone              string                `json:"phone,omitempty"`
	Address            string                `json:"address,omitempty"`
	ContactPerson      string                `json:"contact_person,omitempty"`
	ContactPosition    string                `json:"contact_position,omitempty"`
	ContactTitle       string                `json:"contact_title,omitempty"`
	ContactNote        string                `json:"contact_note,omitempty"`
	RegistrationNumber string                `json:"registration_number,omitempty"`
	CompanyID          *string               `json:"company_id,omitempty"`
	Company            *company.CompanyModel `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
}

func (BrandModel) TableName() string {
	return "brands"
}

func (bm *BrandModel) BeforeCreate(tx *gorm.DB) (err error) {
	bm.ID = uuid.NewString()
	return
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&BrandModel{},
	)
}
