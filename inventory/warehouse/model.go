package warehouse

import (
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// WarehouseModel adalah model database untuk warehouse
type WarehouseModel struct {
	shared.BaseModel
	Name            string `gorm:"not null" json:"name"`
	Code            string `gorm:"type:varchar(255)" json:"code"`
	Description     string `json:"description,omitempty"`
	Address         string `json:"address,omitempty"`
	Phone           string `json:"phone,omitempty"`
	ContactPerson   string `json:"contact_person,omitempty"`
	ContactPosition string `json:"contact_position,omitempty"`
	ContactTitle    string `json:"contact_title,omitempty"`
	ContactNote     string `json:"contact_note,omitempty"`
}

func (WarehouseModel) TableName() string {
	return "warehouses"
}

func (p *WarehouseModel) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&WarehouseModel{})
}
