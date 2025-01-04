package warehouse

import (
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// WarehouseModel adalah model database untuk warehouse
type WarehouseModel struct {
	utils.BaseModel
	Name        string `gorm:"not null"`
	Code        string `gorm:"type:varchar(255)"`
	Description string
	Address     string
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
