package models

import (
	"encoding/json"
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MerchantModel struct {
	shared.BaseModel
	Name                   string              `json:"name" gorm:"not null"`
	Address                string              `json:"address" gorm:"not null"`
	Phone                  string              `json:"phone" gorm:"not null"`
	Latitude               float64             `json:"latitude" gorm:"type:decimal(10,8);not null"`
	Longitude              float64             `json:"longitude" gorm:"type:decimal(11,8);not null"`
	UserID                 *string             `json:"user_id,omitempty" gorm:"index;constraint:OnDelete:CASCADE;"`
	User                   *UserModel          `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;"`
	DefaultWarehouseID     *string             `json:"default_warehouse_id,omitempty" gorm:"type:char(36);index;constraint:OnDelete:CASCADE;"`
	DefaultWarehouse       *WarehouseModel     `json:"default_warehouse,omitempty" gorm:"foreignKey:DefaultWarehouseID;constraint:OnDelete:CASCADE;"`
	DefaultPriceCategoryID *string             `json:"default_price_category_id,omitempty" gorm:"type:char(36);index;constraint:OnDelete:CASCADE;"`
	DefaultPriceCategory   *PriceCategoryModel `json:"default_price_category,omitempty" gorm:"foreignKey:DefaultPriceCategoryID;constraint:OnDelete:CASCADE;"`
	CompanyID              *string             `json:"company_id,omitempty" gorm:"index;constraint:OnDelete:CASCADE;"`
	Company                *CompanyModel       `json:"company,omitempty" gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE;"`
	ProvinceID             *string             `json:"province_id,omitempty" gorm:"type:char(2);index;constraint:OnDelete:SET NULL;"`
	RegencyID              *string             `json:"regency_id,omitempty" gorm:"type:char(4);index;constraint:OnDelete:SET NULL;"`
	DistrictID             *string             `json:"district_id,omitempty" gorm:"type:char(6);index;constraint:OnDelete:SET NULL;"`
	VillageID              *string             `json:"village_id,omitempty" gorm:"type:char(10);index;constraint:OnDelete:SET NULL;"`
	ZipCode                *string             `json:"zip_code,omitempty"`
	Status                 string              `gorm:"type:VARCHAR(20);default:'ACTIVE'" json:"status,omitempty"`
	MerchantType           *string             `json:"merchant_type" gorm:"type:VARCHAR(20);default:'REGULAR_STORE'"`
	MerchantTypeID         *string             `json:"merchant_type_id,omitempty" gorm:"type:char(36);index;constraint:OnDelete:CASCADE;"`
	Picture                *FileModel          `json:"picture,omitempty" gorm:"-"`
	OrderRequest           *OrderRequestModel  `json:"order_request,omitempty" gorm:"-"`
	Distance               float64             `json:"distance" gorm:"-"`
	Users                  []*UserModel        `gorm:"many2many:merchant_users;constraint:OnDelete:CASCADE;" json:"users,omitempty"`
	Workflow               *json.RawMessage    `json:"workflow,omitempty" gorm:"type:JSON;default:'[]'"`
	Menu                   *json.RawMessage    `json:"menu,omitempty" gorm:"type:JSON;default:'[]'"`
	Stations               []MerchantStation   `json:"stations,omitempty" gorm:"foreignKey:MerchantID;constraint:OnDelete:CASCADE;"`
	EnableXendit           bool                `json:"enable_xendit,omitempty" gorm:"default:false"`
	XenditApiKey           string              `json:"xendit_api_key,omitempty" gorm:"type:varchar(255);"`
	XenditApiKeyCensored   string              `json:"xendit_api_key_censored,omitempty" gorm:"-"`
	Xendit                 *XenditModel        `gorm:"foreignKey:MerchantID;constraint:OnDelete:CASCADE;" json:"xendit,omitempty"`
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
	MerchantID           *string             `json:"merchant_id" gorm:"index;constraint:OnDelete:CASCADE;"`
	Merchant             *MerchantModel      `gorm:"foreignKey:MerchantID;constraint:OnDelete:CASCADE;" json:"merchant,omitempty"`
	DeskName             *string             `json:"desk_name"`
	Status               *string             `gorm:"type:varchar(20);default:'AVAILABLE'" json:"status,omitempty"`
	OrderNumber          *int                `json:"order_number"`
	Capacity             int                 `json:"capacity"`
	Position             json.RawMessage     `gorm:"type:JSON;default:'{}'" json:"position"`
	Shape                string              `gorm:"type:varchar(255);default:'rectangle'" json:"shape"`
	Width                float64             `json:"width" gorm:"default:80"`
	Height               float64             `json:"height" gorm:"default:60"`
	MerchantDeskLayoutID *string             `json:"merchant_desk_layout_id" gorm:"index;constraint:OnDelete:CASCADE;"`
	MerchantDeskLayout   *MerchantDeskLayout `gorm:"foreignKey:MerchantDeskLayoutID;constraint:OnDelete:CASCADE;" json:"merchant_desk_layout,omitempty"`
	ContactName          string              `json:"contact_name"`
	ContactPhone         string              `json:"contact_phone"`
	ContactID            *string             `json:"contact_id" gorm:"index;constraint:OnDelete:CASCADE;"`
	Contact              *ContactModel       `gorm:"foreignKey:ContactID;constraint:OnDelete:CASCADE;" json:"contact,omitempty"`
	ActiveOrders         []MerchantOrder     `gorm:"foreignKey:MerchantDeskID;constraint:OnDelete:CASCADE;" json:"active_orders,omitempty"`
}

type MerchantDeskLayout struct {
	shared.BaseModel
	Name          string         `gorm:"type:varchar(255)" json:"name,omitempty"`
	Description   string         `gorm:"type:varchar(255)" json:"description,omitempty"`
	MerchantID    *string        `json:"merchant_id" gorm:"index;constraint:OnDelete:CASCADE;"`
	Merchant      *MerchantModel `gorm:"foreignKey:MerchantID;constraint:OnDelete:CASCADE;" json:"merchant,omitempty"`
	MerchantDesks []MerchantDesk `gorm:"foreignKey:MerchantDeskLayoutID;constraint:OnDelete:CASCADE;" json:"merchant_desks,omitempty"`
}

type MerchantOrder struct {
	shared.BaseModel
	MerchantID            *string                `json:"merchant_id" gorm:"index;constraint:OnDelete:CASCADE;"`
	Merchant              *MerchantModel         `gorm:"foreignKey:MerchantID;constraint:OnDelete:CASCADE;" json:"merchant,omitempty"`
	ContactID             *string                `json:"contact_id" gorm:"index;constraint:OnDelete:CASCADE;"`
	Contact               *ContactModel          `gorm:"foreignKey:ContactID;constraint:OnDelete:CASCADE;" json:"contact,omitempty"`
	ContactData           json.RawMessage        `gorm:"type:JSON;default:'{}'" json:"contact_data,omitempty"`
	Total                 float64                `json:"total,omitempty"`
	SubTotal              float64                `json:"sub_total,omitempty"`
	MerchantDeskID        *string                `json:"merchant_desk_id" gorm:"index;constraint:OnDelete:CASCADE;"`
	MerchantDesk          *MerchantDesk          `gorm:"foreignKey:MerchantDeskID;constraint:OnDelete:CASCADE;" json:"merchant_desk,omitempty"`
	Step                  string                 `json:"step,omitempty"`
	NextStep              string                 `json:"next_step,omitempty" gorm:"-"`
	OrderStatus           string                 `json:"order_status,omitempty"`
	Code                  string                 `json:"code,omitempty"`
	Items                 json.RawMessage        `gorm:"type:JSON;default:'[]'" json:"items,omitempty"`
	MerchantStationOrders []MerchantStationOrder `gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE;" json:"-"`
	MerchantStations      []MerchantStation      `gorm:"-" json:"merchant_stations,omitempty"`
	Payments              []MerchantPayment      `gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE;" json:"payments,omitempty"`
	CashierID             *string                `json:"cashier_id" gorm:"index;constraint:OnDelete:CASCADE;"`
	Cashier               *UserModel             `gorm:"foreignKey:CashierID;constraint:OnDelete:CASCADE;" json:"cashier,omitempty"`
	ContactName           string                 `json:"contact_name" gorm:"-"`
	ContactPhone          string                 `json:"contact_phone" gorm:"-"`
	ParentID              *string                `json:"parent_id" gorm:"index;constraint:OnDelete:CASCADE;"`
	Parent                *MerchantOrder         `gorm:"foreignKey:ParentID;constraint:OnDelete:CASCADE;" json:"parent,omitempty"`
}

type MerchantOrderItem struct {
	ID                 string       `json:"id,omitempty"`
	ProductID          string       `json:"product_id,omitempty"`
	Product            ProductModel `json:"product,omitempty"`
	Quantity           float64      `json:"quantity,omitempty"`
	DiscountAmount     float64      `json:"discount_amount,omitempty"`
	DiscountPercent    float64      `json:"discount_percent,omitempty"`
	UnitPrice          float64      `json:"unit_price,omitempty"`
	SubtotalBeforeDisc float64      `json:"subtotal_before_disc,omitempty"`
	Subtotal           float64      `json:"subtotal,omitempty"`
	UnitName           string       `json:"unit_name,omitempty"`
	UnitValue          float64      `json:"unit_value,omitempty"`
	Notes              string       `json:"notes,omitempty"`
}
type MerchantStation struct {
	shared.BaseModel
	MerchantID  *string                `json:"merchant_id" gorm:"index;constraint:OnDelete:CASCADE;"`
	Merchant    *MerchantModel         `gorm:"foreignKey:MerchantID;constraint:OnDelete:CASCADE;" json:"merchant,omitempty"`
	StationName string                 `json:"station_name"`
	Description string                 `json:"description"`
	Orders      []MerchantStationOrder `gorm:"foreignKey:MerchantStationID;constraint:OnDelete:CASCADE;" json:"orders,omitempty"`
	Products    []ProductModel         `gorm:"-" json:"products,omitempty"`
}

type MerchantStationOrder struct {
	shared.BaseModel
	MerchantStationID *string          `json:"merchant_station_id" gorm:"index;constraint:OnDelete:CASCADE;"`
	MerchantStation   *MerchantStation `gorm:"foreignKey:MerchantStationID;constraint:OnDelete:CASCADE;" json:"merchant_station,omitempty"`
	OrderID           string           `json:"order_id"`
	Order             *MerchantOrder   `gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE;" json:"order,omitempty"`
	Status            string           `json:"status,omitempty"`
	Item              json.RawMessage  `gorm:"type:JSON;default:'{}'" json:"item,omitempty"`
	MerchantDeskID    *string          `json:"merchant_desk_id" gorm:"index;constraint:OnDelete:CASCADE;"`
	MerchantDesk      *MerchantDesk    `gorm:"foreignKey:MerchantDeskID;constraint:OnDelete:CASCADE;" json:"merchant_desk,omitempty"`
}

type MerchantPayment struct {
	shared.BaseModel
	Date             time.Time       `json:"date"`
	MerchantID       *string         `json:"merchant_id" gorm:"index;constraint:OnDelete:CASCADE;"`
	Merchant         *MerchantModel  `gorm:"foreignKey:MerchantID;constraint:OnDelete:CASCADE;" json:"merchant,omitempty"`
	OrderID          string          `json:"order_id"`
	Order            *MerchantOrder  `gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE;" json:"order,omitempty"`
	Amount           float64         `json:"amount"`
	Change           float64         `json:"change"`
	Notes            string          `json:"notes"`
	PaymentMethod    string          `gorm:"type:varchar(255)" json:"payment_method"`
	PaymentProvider  string          `gorm:"type:varchar(255)" json:"payment_provider"`
	ExternalID       string          `gorm:"type:varchar(255)" json:"external_id"`
	ExternalRef      string          `gorm:"type:varchar(255)" json:"external_ref"`
	ExternalProvider string          `gorm:"type:varchar(255)" json:"external_provider"`
	ExternalURL      string          `gorm:"type:varchar(255)" json:"external_url"`
	PaymentData      json.RawMessage `gorm:"type:JSON;default:'{}'" json:"payment_data"`
}
