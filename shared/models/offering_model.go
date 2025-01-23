package models

import (
	"encoding/json"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OfferModel struct {
	shared.BaseModel
	UserID                       string                   `json:"user_id"`
	OrderRequestID               string                   `json:"order_request_id"`
	OrderRequest                 OrderRequestModel        `gorm:"foreignKey:OrderRequestID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"order_request,omitempty"`
	MerchantID                   string                   `json:"merchant_id"`
	Merchant                     MerchantModel            `gorm:"foreignKey:MerchantID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"merchant,omitempty"`
	PaymentID                    *string                  `json:"payment_id"`
	Payment                      *PaymentModel            `gorm:"foreignKey:PaymentID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"payment,omitempty"`
	SubTotal                     float64                  `json:"sub_total"`
	TotalPrice                   float64                  `json:"total_price"`
	ShippingFee                  float64                  `json:"shipping_fee"`
	Distance                     float64                  `json:"distance"`
	Status                       string                   `json:"status"` // Pending, Accepted, Taken
	MerchantAvailableProduct     MerchantAvailableProduct `json:"merchant_available_product" gorm:"-"`
	MerchantAvailableProductData string                   `json:"-" gorm:"type:json"`
}

func (OfferModel) TableName() string {
	return "offers"
}

func (o *OfferModel) BeforeCreate(tx *gorm.DB) (err error) {
	o.Status = "PENDING"
	if o.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}
func (o *OfferModel) AfterFind(tx *gorm.DB) (err error) {
	if o.MerchantAvailableProductData != "" {
		if err = json.Unmarshal([]byte(o.MerchantAvailableProductData), &o.MerchantAvailableProduct); err != nil {
			return err
		}
	}
	return
}
