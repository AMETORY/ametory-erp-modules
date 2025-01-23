package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrderRequestModel struct {
	shared.BaseModel
	UserID             string                  `json:"user_id,omitempty"`
	User               UserModel               `gorm:"foreignKey:UserID" json:"user,omitempty"`
	ContactID          *string                 `json:"contact_id,omitempty"`
	Contact            *ContactModel           `gorm:"foreignKey:ContactID" json:"contact,omitempty"`
	UserLat            float64                 `json:"user_lat,omitempty"`
	UserLng            float64                 `json:"user_lng,omitempty"`
	Status             string                  `json:"status,omitempty"`                                 // "Pending", "Accepted", "Rejected"
	MerchantID         *string                 `gorm:"type:char(36);index" json:"merchant_id,omitempty"` // Diisi jika merchant mengambil order
	TotalPrice         float64                 `json:"total_price,omitempty"`
	SubTotal           float64                 `json:"sub_total,omitempty"`
	ShippingFee        float64                 `json:"shipping_fee,omitempty"`
	ShippingData       string                  `gorm:"type:json" json:"shipping_data,omitempty"`
	Distance           float64                 `json:"distance"`
	ExpiresAt          time.Time               `json:"expires_at,omitempty"` // Batas waktu pengambilan order
	Items              []OrderRequestItemModel `gorm:"foreignKey:OrderRequestID;constraint:OnDelete:CASCADE" json:"items"`
	Offers             []OfferModel            `gorm:"foreignKey:OrderRequestID;constraint:OnDelete:CASCADE" json:"offers,omitempty"`
	CancellationReason string                  `json:"cancellation_reason,omitempty"`
}

func (OrderRequestModel) TableName() string {
	return "order_requests"
}

func (orm *OrderRequestModel) BeforeCreate(tx *gorm.DB) (err error) {
	orm.ID = uuid.New().String()
	return
}

// OrderRequestItemModel adalah representasi di database untuk item order request
type OrderRequestItemModel struct {
	shared.BaseModel
	OrderRequestID  string        `gorm:"type:char(36);index" json:"-"`
	Description     string        `json:"description"`
	Quantity        float64       `json:"quantity"`
	UnitPrice       float64       `json:"unit_price"`
	DiscountPercent float64       `json:"discount_percent"`
	DiscountAmount  float64       `json:"discount_amount"`
	Total           float64       `json:"total"`
	ProductID       *string       `json:"product_id"`
	Product         *ProductModel `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	VariantID       *string       `json:"variant_id"`
	Variant         *VariantModel `gorm:"foreignKey:VariantID" json:"variant,omitempty"`
	Status          string        `json:"status" gorm:"-"`
}

func (OrderRequestItemModel) TableName() string {
	return "order_request_items"
}

func (orim *OrderRequestItemModel) BeforeCreate(tx *gorm.DB) (err error) {
	orim.ID = uuid.New().String()
	return
}
