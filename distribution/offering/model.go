package offering

import (
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OfferModel struct {
	shared.BaseModel
	UserID         string  `json:"user_id"`
	OrderRequestID string  `json:"order_request_id"`
	MerchantID     string  `json:"merchant_id"`
	TotalPrice     float64 `json:"total_price"`
	ShippingFee    float64 `json:"shipping_fee"`
	Distance       float64 `json:"distance"`
	Status         string  `json:"status"` // Pending, Accepted, Taken
}

func (OfferModel) TableName() string {
	return "offers"
}

func (o *OfferModel) BeforeCreate(tx *gorm.DB) (err error) {
	o.Status = "Pending"
	if o.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}

func Migrate(tx *gorm.DB) error {
	return tx.AutoMigrate(&OfferModel{})
}
