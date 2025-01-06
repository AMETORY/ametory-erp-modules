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
	Name               string                `gorm:"not null" json:"name"`
	Description        string                `json:"description"`
	Logo               string                `json:"logo"`
	Website            string                `json:"website"`
	Email              string                `json:"email"`
	Phone              string                `json:"phone"`
	Address            string                `json:"address"`
	ContactPerson      string                `json:"contact_person"`
	ContactPosition    string                `json:"contact_position"`
	ContactTitle       string                `json:"contact_title"`
	ContactNote        string                `json:"contact_note"`
	RegistrationNumber string                `json:"registration_number"`
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
