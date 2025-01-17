package merchant

import (
	"github.com/AMETORY/ametory-erp-modules/company"
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MerchantModel struct {
	shared.BaseModel
	Name               string                `json:"name" gorm:"not null"`
	Address            string                `json:"address" gorm:"not null"`
	Phone              string                `json:"phone" gorm:"not null"`
	Latitude           float64               `json:"latitude" gorm:"type:decimal(10,8);not null"`
	Longitude          float64               `json:"longitude" gorm:"type:decimal(11,8);not null"`
	UserID             *string               `json:"user_id,omitempty" gorm:"index;constraint:OnDelete:CASCADE;"`
	CompanyID          *string               `json:"company_id,omitempty" gorm:"index;constraint:OnDelete:CASCADE;"`
	DefaultWarehouseID *string               `json:"default_warehouse_id,omitempty" gorm:"type:char(36);index"`
	Company            *company.CompanyModel `json:"company,omitempty" gorm:"foreignKey:CompanyID"`
	ProvinceID         *string               `json:"province_id,omitempty" gorm:"type:char(2);index"`
	RegencyID          *string               `json:"regency_id,omitempty" gorm:"type:char(4);index"`
	DistrictID         *string               `json:"district_id,omitempty" gorm:"type:char(6);index"`
	VillageID          *string               `json:"village_id,omitempty" gorm:"type:char(10);index"`
	Status             string                `gorm:"type:VARCHAR(20);default:'ACTIVE'" json:"status,omitempty"`
	MerchantType       string                `json:"merchant_type" gorm:"type:VARCHAR(20);default:'REGULAR_STORE'"`
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
