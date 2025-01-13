package merchant

import (
	"github.com/AMETORY/ametory-erp-modules/company"
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MerchantModel struct {
	shared.BaseModel
	Name      string                `json:"name" gorm:"not null"`
	Address   string                `json:"address" gorm:"not null"`
	Phone     string                `json:"phone" gorm:"not null"`
	Latitude  float64               `json:"latitude" gorm:"type:decimal(10,8);not null"`
	Longitude float64               `json:"longitude" gorm:"type:decimal(11,8);not null"`
	CompanyID *string               `json:"company_id,omitempty" gorm:"index"`
	Company   *company.CompanyModel `json:"company,omitempty" gorm:"foreignKey:CompanyID"`
}

func (m *MerchantModel) TableName() string {
	return "pos_merchants"
}

func (m *MerchantModel) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&MerchantModel{})
}
