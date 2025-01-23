package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PaymentProviderType string

const (
	CREDIT_CARD PaymentProviderType = "CREDIT_CARD"
	PAYPAL      PaymentProviderType = "PAYPAL"
	BANK        PaymentProviderType = "BANK"
	CASH        PaymentProviderType = "CASH"
	NON_CASH    PaymentProviderType = "NON_CASH"
	MULTIPLE    PaymentProviderType = "MULTIPLE"
	BCA         PaymentProviderType = "BCA"
	MANDIRI     PaymentProviderType = "MANDIRI"
	BRI         PaymentProviderType = "BRI"
	BNI         PaymentProviderType = "BNI"
	CIMB        PaymentProviderType = "CIMB"
	SHOPEE      PaymentProviderType = "SHOPEE"
	OVO         PaymentProviderType = "OVO"
	GOPAY       PaymentProviderType = "GOPAY"
	DANA        PaymentProviderType = "DANA"
	LINKAJA     PaymentProviderType = "LINKAJA"
	GIFTCARD    PaymentProviderType = "GIFTCARD"
	GOFOOD      PaymentProviderType = "GOFOOD"
	GRABFOOD    PaymentProviderType = "GRABFOOD"
	QRIS        PaymentProviderType = "QRIS"
	OTHER       PaymentProviderType = "OTHER"
)

type POSModel struct {
	shared.BaseModel
	SalesNumber         string              `json:"sales_number" gorm:"sales_number"`
	Code                string              `json:"code" gorm:"code"`
	Description         string              `json:"description" gorm:"description"`
	Notes               string              `json:"notes" gorm:"notes"`
	Total               float64             `json:"total" gorm:"total"`
	Subtotal            float64             `json:"subtotal" gorm:"subtotal"`
	ShippingFee         float64             `json:"shipping_fee" gorm:"shipping_fee"`
	Paid                float64             `json:"paid" gorm:"paid"`
	TotalBeforeTax      float64             `json:"total_before_tax" gorm:"total_before_tax"`
	TotalBeforeDisc     float64             `json:"total_before_disc" gorm:"total_before_disc"`
	Status              string              `json:"status" gorm:"status"`
	StockStatus         string              `json:"stock_status" gorm:"default:'pending';stock_status"`
	SalesDate           time.Time           `json:"sales_date" gorm:"sales_date"`
	DueDate             time.Time           `json:"due_date" gorm:"due_date"`
	PaymentTerms        string              `json:"payment_terms" gorm:"payment_terms"`
	PaymentID           *string             `json:"payment_id" gorm:"payment_id"`
	Payment             *PaymentModel       `gorm:"foreignKey:PaymentID;constraint:OnDelete:CASCADE"`
	OfferID             *string             `json:"offer_id" gorm:"offer_id"`
	Offer               *OfferModel         `gorm:"foreignKey:OfferID;constraint:OnDelete:CASCADE"`
	MerchantID          *string             `json:"merchant_id" gorm:"merchant_id"`
	Merchant            *MerchantModel      `gorm:"foreignKey:MerchantID;constraint:OnDelete:CASCADE"`
	CompanyID           *string             `json:"company_id" gorm:"company_id"`
	Company             *CompanyModel       `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE"`
	ContactID           *string             `json:"contact_id" gorm:"contact_id"`
	Contact             *ContactModel       `gorm:"foreignKey:ContactID;constraint:OnDelete:CASCADE"`
	ContactData         string              `json:"contact_data" gorm:"type:json;contact_data"`
	PaymentType         string              `json:"payment_type" gorm:"payment_type"`
	PaymentProviderType PaymentProviderType `json:"payment_provider_type" gorm:"payment_provider_type"`
	Items               []POSSalesItemModel `json:"items" gorm:"foreignKey:SalesID;constraint:OnDelete:CASCADE"`
	SaleAccountID       *string             `json:"sale_account_id" gorm:"sale_account_id"`
	SaleAccount         *AccountModel       `gorm:"foreignKey:SaleAccountID;constraint:OnDelete:CASCADE"`
	AssetAccountID      *string             `json:"asset_account_id" gorm:"asset_account_id"`
	AssetAccount        *AccountModel       `gorm:"foreignKey:AssetAccountID;constraint:OnDelete:CASCADE"`
}

type POSSalesItemModel struct {
	shared.BaseModel
	SalesID            string          `json:"sales_id"`
	Sales              POSModel        `gorm:"foreignKey:SalesID;constraint:OnDelete:CASCADE"`
	Description        string          `json:"description"`
	Quantity           float64         `json:"quantity"`
	UnitPrice          float64         `json:"unit_price"`
	Total              float64         `json:"total"`
	DiscountPercent    float64         `json:"discount_percent"`
	DiscountAmount     float64         `json:"discount_amount"`
	SubtotalBeforeDisc float64         `json:"subtotal_before_disc"`
	ProductID          *string         `json:"product_id"`
	Product            *ProductModel   `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE"`
	VariantID          *string         `json:"variant_id,omitempty"`
	Variant            *VariantModel   `gorm:"foreignKey:VariantID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
	WarehouseID        *string         `json:"warehouse_id"`
	Warehouse          *WarehouseModel `gorm:"foreignKey:WarehouseID;constraint:OnDelete:CASCADE"`
}

func (s *POSModel) TableName() string {
	return "pos_sales"
}

func (s *POSModel) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}

func (s *POSSalesItemModel) TableName() string {
	return "pos_sales_items"
}

func (s *POSSalesItemModel) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}
