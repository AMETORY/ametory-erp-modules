package order_request

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/auth"
	"github.com/AMETORY/ametory-erp-modules/shared"
)

type OrderRequestModel struct {
	shared.BaseModel
	UserID             string                  `json:"user_id,omitempty"`
	User               auth.UserModel          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	UserLat            float64                 `json:"user_lat,omitempty"`
	UserLng            float64                 `json:"user_lng,omitempty"`
	Status             string                  `json:"status,omitempty"`                                 // "Pending", "Accepted", "Rejected"
	MerchantID         *string                 `gorm:"type:char(36);index" json:"merchant_id,omitempty"` // Diisi jika merchant mengambil order
	TotalPrice         float64                 `json:"total_price,omitempty"`
	ShippingFee        float64                 `json:"shipping_fee,omitempty"`
	ExpiresAt          time.Time               `json:"expires_at,omitempty"` // Batas waktu pengambilan order
	Items              []OrderRequestItemModel `gorm:"foreignKey:OrderRequestID" json:"items,omitempty"`
	CancellationReason string                  `json:"cancellation_reason,omitempty"`
}

// OrderRequestItemModel adalah representasi di database untuk item order request
type OrderRequestItemModel struct {
	shared.BaseModel
	OrderRequestID string  `gorm:"type:char(36);index" json:"-"`
	Description    string  `json:"description"`
	Quantity       float64 `json:"quantity"`
	UnitPrice      float64 `json:"unit_price"`
	Total          float64 `json:"total"`
	ProductID      *string `json:"product_id"`
}
