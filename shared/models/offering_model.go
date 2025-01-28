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
	OrderRequest                 *OrderRequestModel       `gorm:"foreignKey:OrderRequestID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"order_request,omitempty"`
	MerchantID                   string                   `json:"merchant_id"`
	Merchant                     *MerchantModel           `gorm:"foreignKey:MerchantID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"merchant,omitempty"`
	PaymentID                    *string                  `json:"payment_id"`
	Payment                      *PaymentModel            `gorm:"foreignKey:PaymentID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"payment,omitempty"`
	SubTotal                     float64                  `json:"sub_total"`
	SubTotalBeforeDiscount       float64                  `json:"sub_total_before_discount"`
	TotalDiscountAmount          float64                  `json:"total_discount_amount"`
	TotalPrice                   float64                  `json:"total_price"`
	ShippingFee                  float64                  `json:"shipping_fee"`
	ServiceFee                   float64                  `json:"service_fee"`
	Tax                          float64                  `json:"tax"`
	TaxType                      string                   `json:"tax_type" gorm:"type:varchar"`
	TaxAmount                    float64                  `json:"tax_amount"`
	TotalTaxAmount               float64                  `json:"total_tax_amount"`
	ShippingType                 string                   `json:"service_type" gorm:"type:varchar(50)"`
	CourierName                  string                   `json:"courier_name"`
	Distance                     float64                  `json:"distance"`
	Status                       string                   `json:"status"` // Pending, Accepted, Taken
	MerchantAvailableProduct     MerchantAvailableProduct `json:"merchant_available_product" gorm:"-"`
	MerchantAvailableProductData string                   `json:"-" gorm:"type:json"`
	SubTotalAfterDiscount        float64                  `json:"sub_total_after_discount"`
	DiscountAmount               float64                  `json:"discount_amount"`
	DiscountValue                float64                  `json:"discount_value"`
	DiscountType                 string                   `json:"discount_type"`
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
