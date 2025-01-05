package pos

import (
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MerchantModel struct {
	utils.BaseModel
	Name    string
	Address string
	Phone   string
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
