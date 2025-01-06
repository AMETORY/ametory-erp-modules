package distributor

import (
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DistributorModel adalah model database untuk distributor
type DistributorModel struct {
	utils.BaseModel
	Name    string `json:"name"`
	Address string `json:"address"`
	Phone   string `json:"phone"`
}

func (DistributorModel) TableName() string {
	return "distributors"
}

func (m *DistributorModel) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	return
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&DistributorModel{})
}
