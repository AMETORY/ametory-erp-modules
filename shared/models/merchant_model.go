package models

import (
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MerchantModel struct {
	shared.BaseModel
	Name               string             `json:"name" gorm:"not null"`
	Address            string             `json:"address" gorm:"not null"`
	Phone              string             `json:"phone" gorm:"not null"`
	Latitude           float64            `json:"latitude" gorm:"type:decimal(10,8);not null"`
	Longitude          float64            `json:"longitude" gorm:"type:decimal(11,8);not null"`
	UserID             *string            `json:"user_id,omitempty" gorm:"index;constraint:OnDelete:CASCADE;"`
	User               *UserModel         `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;"`
	CompanyID          *string            `json:"company_id,omitempty" gorm:"index;constraint:OnDelete:CASCADE;"`
	DefaultWarehouseID *string            `json:"default_warehouse_id,omitempty" gorm:"type:char(36);index;constraint:OnDelete:CASCADE;"`
	DefaultWarehouse   *WarehouseModel    `json:"default_warehouse,omitempty" gorm:"foreignKey:DefaultWarehouseID;constraint:OnDelete:CASCADE;"`
	Company            *CompanyModel      `json:"company,omitempty" gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE;"`
	ProvinceID         *string            `json:"province_id,omitempty" gorm:"type:char(2);index;constraint:OnDelete:SET NULL;"`
	RegencyID          *string            `json:"regency_id,omitempty" gorm:"type:char(4);index;constraint:OnDelete:SET NULL;"`
	DistrictID         *string            `json:"district_id,omitempty" gorm:"type:char(6);index;constraint:OnDelete:SET NULL;"`
	VillageID          *string            `json:"village_id,omitempty" gorm:"type:char(10);index;constraint:OnDelete:SET NULL;"`
	Status             string             `gorm:"type:VARCHAR(20);default:'ACTIVE'" json:"status,omitempty"`
	MerchantType       string             `json:"merchant_type" gorm:"type:VARCHAR(20);default:'REGULAR_STORE'"`
	Picture            *FileModel         `json:"picture,omitempty" gorm:"-"`
	OrderRequest       *OrderRequestModel `json:"order_request,omitempty" gorm:"-"`
	Distance           float64            `json:"distance" gorm:"-"`
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

type MerchantAvailableProduct struct {
	Name           string                         `json:"name" gorm:"not null"`
	MerchantID     string                         `json:"merchant_id"`
	OrderRequestID string                         `json:"order_request_id"`
	Items          []MerchantAvailableProductItem `json:"items"`
	SubTotal       float64                        `json:"sub_total"`
	ShippingFee    float64                        `json:"shipping_fee"`
	TotalPrice     float64                        `json:"total_price"`
	Distance       float64                        `json:"distance"`
}

type MerchantAvailableProductItem struct {
	ProductID          string  `json:"product_id"`
	ProductDisplayName string  `json:"product_display_name"`
	VariantDisplayName *string `json:"variant_display_name"`
	Status             string  `json:"status"`
	VariantID          *string `json:"variant_id"`
	Quantity           float64 `json:"quantity"`
	UnitPrice          float64 `json:"unit_price"`
	SubTotal           float64 `json:"sub_total"`
}
