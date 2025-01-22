package models

import (
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CartModel struct {
	shared.BaseModel
	Code     string          `json:"code,omitempty"`
	UserID   string          `gorm:"not null" json:"user_id,omitempty"`
	User     UserModel       `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	Items    []CartItemModel `gorm:"foreignKey:CartID;constraint:OnDelete:CASCADE" json:"items,omitempty"`
	Status   string          `gorm:"type:varchar(50);not null;default:'ACTIVE'" json:"status,omitempty"`
	SubTotal float64         `gorm:"-" json:"sub_total,omitempty"`
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
	CartID    string        `gorm:"not null" json:"cart_id,omitempty"`
	Cart      *CartModel    `gorm:"foreignKey:CartID;constraint:OnDelete:CASCADE" json:"cart,omitempty"`
	ProductID string        `gorm:"not null" json:"product_id,omitempty"`
	Product   ProductModel  `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE" json:"product,omitempty"`
	VariantID *string       `json:"variant_id,omitempty"`
	Variant   *VariantModel `gorm:"foreignKey:VariantID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
	Quantity  float64       `gorm:"not null" json:"quantity,omitempty"`
	Price     float64       `gorm:"not null" json:"price,omitempty"`
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
