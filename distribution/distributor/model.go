package distributor

import (
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DistributorModel adalah model database untuk distributor
type DistributorModel struct {
	shared.BaseModel
	Name            string `json:"name"`
	Address         string `json:"address"`
	Phone           string `json:"phone"`
	Website         string `json:"website"`
	Email           string `json:"email"`
	ContactPerson   string `json:"contact_person"`
	ContactPosition string `json:"contact_position"`
	ContactTitle    string `json:"contact_title"`
	ContactNote     string `json:"contact_note"`
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
