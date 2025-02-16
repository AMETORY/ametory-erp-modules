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
	User               UserModel               `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	ContactID          *string                 `json:"contact_id,omitempty"`
	Contact            *ContactModel           `gorm:"foreignKey:ContactID;constraint:OnDelete:CASCADE" json:"contact,omitempty"`
	UserLat            float64                 `json:"user_lat,omitempty"`
	UserLng            float64                 `json:"user_lng,omitempty"`
	Status             string                  `json:"status,omitempty"`                                 // "Pending", "Accepted", "Rejected"
	MerchantID         *string                 `gorm:"type:char(36);index" json:"merchant_id,omitempty"` // Diisi jika merchant mengambil order
	TotalPrice         float64                 `json:"total_price,omitempty"`
	SubTotal           float64                 `json:"sub_total,omitempty"`
	ShippingFee        float64                 `json:"shipping_fee,omitempty"`
	ServiceFee         float64                 `json:"service_fee,omitempty"`
	Tax                float64                 `json:"tax"`
	TaxType            string                  `json:"tax_type" gorm:"type:varchar"`
	TaxAmount          float64                 `json:"tax_amount"`
	TotalTaxAmount     float64                 `json:"total_tax_amount"`
	ShippingData       string                  `gorm:"type:json" json:"-"`
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
	OrderRequestID  string                `gorm:"type:char(36);index" json:"-"`
	Description     string                `json:"description"`
	Quantity        float64               `json:"quantity"`
	UnitPrice       float64               `json:"unit_price"`
	DiscountPercent float64               `json:"discount_percent"`
	DiscountAmount  float64               `json:"discount_amount"`
	Total           float64               `json:"total"`
	ProductID       *string               `json:"product_id"`
	Product         *ProductModel         `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	VariantID       *string               `json:"variant_id"`
	Variant         *VariantModel         `gorm:"foreignKey:VariantID" json:"variant,omitempty"`
	Status          string                `json:"status" gorm:"-"`
	CategoryID      *string               `json:"category_id,omitempty"`
	Category        *ProductCategoryModel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:CategoryID" json:"category,omitempty"`
	Brand           *BrandModel           `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:BrandID" json:"brand,omitempty"`
	BrandID         *string               `json:"brand_id,omitempty"`
	ProductImages   []FileModel           `gorm:"-" json:"product_images,omitempty"`
}

func (OrderRequestItemModel) TableName() string {
	return "order_request_items"
}

func (orim *OrderRequestItemModel) BeforeCreate(tx *gorm.DB) (err error) {
	orim.ID = uuid.New().String()
	return
}
