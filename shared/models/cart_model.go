package models

import (
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CartModel struct {
	shared.BaseModel
	Code                   string          `json:"code,omitempty"`
	UserID                 string          `gorm:"not null" json:"user_id,omitempty"`
	User                   UserModel       `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	MerchantID             *string         `json:"merchant_id,omitempty" gorm:"column:merchant_id"`
	Merchant               *MerchantModel  `gorm:"foreignKey:MerchantID;constraint:OnDelete:CASCADE" json:"merchant,omitempty"`
	Items                  []CartItemModel `gorm:"foreignKey:CartID;constraint:OnDelete:CASCADE" json:"items,omitempty"`
	Status                 string          `gorm:"type:varchar(50);not null;default:'ACTIVE'" json:"status,omitempty"`
	SubTotal               float64         `gorm:"-" json:"sub_total,omitempty"`
	SubTotalBeforeDiscount float64         `gorm:"-" json:"sub_total_before_discount,omitempty"`
	Total                  float64         `gorm:"-" json:"total,omitempty"`
	TaxAmount              float64         `gorm:"-" json:"tax_amount,omitempty"`
	DiscountAmount         float64         `gorm:"-" json:"discount_amount,omitempty"`
	CustomerData           string          `gorm:"-" json:"-"`
	CustomerDataResponse   interface{}     `gorm:"-" json:"customer_data_response,omitempty"`
	Tax                    float64         `gorm:"-" json:"tax"`
	ServiceFee             float64         `gorm:"-" json:"service_fee"`
	TaxType                string          `gorm:"-" json:"tax_type"`
}

func (CartModel) TableName() string {
	return "carts"
}
func (c *CartModel) BeforeCreate(tx *gorm.DB) (err error) {
	c.Code = utils.RandString(7, true)
	// Ensure UserID is not empty
	if c.ID == "" {
		c.ID = uuid.NewString()
	}
	return nil
}

type CartItemModel struct {
	shared.BaseModel
	CartID                 string                `gorm:"not null" json:"cart_id,omitempty"`
	Cart                   *CartModel            `gorm:"foreignKey:CartID;constraint:OnDelete:CASCADE" json:"cart,omitempty"`
	ProductID              string                `gorm:"not null" json:"product_id,omitempty"`
	Product                ProductModel          `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE" json:"-"`
	VariantID              *string               `json:"variant_id,omitempty"`
	Variant                *VariantModel         `gorm:"foreignKey:VariantID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
	Quantity               float64               `gorm:"not null" json:"quantity,omitempty"`
	Price                  float64               `gorm:"not null" json:"price,omitempty"`
	OriginalPrice          float64               `gorm:"-" json:"original_price,omitempty"`
	DiscountAmount         float64               `gorm:"-" json:"discount_amount,omitempty"`
	DiscountType           string                `gorm:"-" json:"discount_type,omitempty"`
	DiscountRate           float64               `gorm:"-" json:"discount_rate,omitempty"`
	AdjustmentPrice        float64               `gorm:"-" json:"adjustment_price,omitempty"`
	ActiveDiscount         *DiscountModel        `gorm:"-" json:"active_discount,omitempty"`
	DisplayName            string                `gorm:"-" json:"display_name,omitempty"`
	ProductImages          []FileModel           `gorm:"-" json:"product_images,omitempty"`
	Height                 float64               `gorm:"default:10" json:"height,omitempty"`
	Length                 float64               `gorm:"default:10" json:"length,omitempty"`
	Weight                 float64               `gorm:"default:200" json:"weight,omitempty"`
	Width                  float64               `gorm:"default:10" json:"width,omitempty"`
	SubTotal               float64               `gorm:"-" json:"sub_total,omitempty"`
	SubTotalBeforeDiscount float64               `gorm:"-" json:"sub_total_before_discount,omitempty"`
	CategoryID             *string               `json:"category_id,omitempty"`
	Category               *ProductCategoryModel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:CategoryID" json:"category,omitempty"`
	Brand                  *BrandModel           `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:BrandID" json:"brand,omitempty"`
	BrandID                *string               `json:"brand_id,omitempty"`
}

func (CartItemModel) TableName() string {
	return "cart_items"
}

func (c *CartItemModel) BeforeCreate(tx *gorm.DB) (err error) {
	if c.ID == "" {
		c.ID = uuid.NewString()
	}
	return nil
}
