package models

import (
	"encoding/json"

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
	ZipCode            *string            `json:"zip_code,omitempty"`
	Status             string             `gorm:"type:VARCHAR(20);default:'ACTIVE'" json:"status,omitempty"`
	MerchantType       *string            `json:"merchant_type" gorm:"type:VARCHAR(20);default:'REGULAR_STORE'"`
	MerchantTypeID     *string            `json:"merchant_type_id,omitempty" gorm:"type:char(36);index;constraint:OnDelete:CASCADE;"`
	Picture            *FileModel         `json:"picture,omitempty" gorm:"-"`
	OrderRequest       *OrderRequestModel `json:"order_request,omitempty" gorm:"-"`
	Distance           float64            `json:"distance" gorm:"-"`
	Users              []*UserModel       `gorm:"many2many:merchant_users;constraint:OnDelete:CASCADE;" json:"users,omitempty"`
	Workflow           *json.RawMessage   `json:"workflow,omitempty" gorm:"type:JSON;default:'[]'"`
	Menu               *json.RawMessage   `json:"menu,omitempty" gorm:"type:JSON;default:'[]'"`
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
	Name                   string                         `json:"name" gorm:"not null"`
	MerchantID             string                         `json:"merchant_id"`
	OrderRequestID         string                         `json:"order_request_id"`
	Items                  []MerchantAvailableProductItem `json:"items"`
	SubTotal               float64                        `json:"sub_total"`
	SubTotalBeforeDiscount float64                        `json:"sub_total_before_discount"`
	ShippingFee            float64                        `json:"shipping_fee"`
	ShippingType           string                         `json:"service_type"`
	Tax                    float64                        `json:"tax"`
	TaxType                string                         `json:"tax_type" gorm:"type:varchar"`
	TaxAmount              float64                        `json:"tax_amount"`
	TotalTaxAmount         float64                        `json:"total_tax_amount"`
	TotalDiscountAmount    float64                        `json:"total_discount_amount"`
	CourierName            string                         `json:"courier_name"`
	ServiceFee             float64                        `json:"service_fee"`
	TotalPrice             float64                        `json:"total_price"`
	Distance               float64                        `json:"distance"`
	// SecondaryShippingFee   float64                        `json:"secondary_shipping_fee"`
	// SecondaryShippingType  string                         `json:"secondary_service_type"`
	// SecondaryCourierName   string                         `json:"secondary_courier_name"`
}

type MerchantAvailableProductItem struct {
	ProductID               string      `json:"product_id"`
	ProductDisplayName      string      `json:"product_display_name"`
	VariantDisplayName      *string     `json:"variant_display_name"`
	Status                  string      `json:"status"`
	VariantID               *string     `json:"variant_id"`
	Quantity                float64     `json:"quantity"`
	UnitPrice               float64     `json:"unit_price"`
	UnitPriceBeforeDiscount float64     `json:"unit_price_before_discount"`
	SubTotal                float64     `json:"sub_total"`
	SubTotalBeforeDiscount  float64     `json:"sub_total_before_discount"`
	DiscountAmount          float64     `json:"discount_amount"`
	DiscountValue           float64     `json:"discount_value"`
	DiscountType            string      `json:"discount_type"`
	ProductImages           []FileModel `gorm:"-" json:"product_images,omitempty"`
}

type MerchantTypeModel struct {
	shared.BaseModel
	Name                string `gorm:"type:varchar(20);not null" json:"name,omitempty"`
	Description         string `gorm:"type:varchar(255)" json:"description,omitempty"`
	IconURL             string `gorm:"type:varchar(255)" json:"icon_url,omitempty"`
	IconBackgroundColor string `gorm:"type:varchar(20)" json:"icon_background_color,omitempty"`
}

func (m *MerchantTypeModel) TableName() string {
	return "pos_merchant_types"
}
func (m *MerchantTypeModel) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	return
}

func (m *MerchantAvailableProductItem) AfterFind(tx *gorm.DB) (err error) {
	var images []FileModel
	tx.Where("ref_id = ? and ref_type = ?", m.ProductID, "product").Find(&images)
	m.ProductImages = images
	return
}

type MerchantUser struct {
	UserModelID     string `gorm:"primaryKey;uniqueIndex:merchant_users_user_id_merchant_id_key" json:"user_id"`
	MerchantModelID string `gorm:"primaryKey;uniqueIndex:merchant_users_user_id_merchant_id_key" json:"merchant_id"`
}

type MerchantDesk struct {
	shared.BaseModel
	MerchantID *string        `json:"merchant_id" gorm:"index;constraint:OnDelete:CASCADE;"`
	Merchant   *MerchantModel `gorm:"foreignKey:MerchantID;constraint:OnDelete:CASCADE;" json:"merchant,omitempty"`
	DeskName   string         `json:"desk_name"`
	Status     string         `gorm:"type:varchar(20);default:'AVAILABLE'" json:"status,omitempty"`
	Position   int            `json:"position"`
	Capacity   int            `json:"capacity"`
}
