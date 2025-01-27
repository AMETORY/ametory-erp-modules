package models

import (
	"encoding/json"
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
	SalesNumber         string                 `json:"sales_number,omitempty" gorm:"column:sales_number"`
	Code                string                 `json:"code,omitempty" gorm:"column:code"`
	Description         string                 `json:"description,omitempty" gorm:"column:description"`
	Notes               string                 `json:"notes,omitempty" gorm:"column:notes"`
	Total               float64                `json:"total,omitempty" gorm:"column:total"`
	Subtotal            float64                `json:"subtotal,omitempty" gorm:"column:subtotal"`
	ShippingFee         float64                `json:"shipping_fee,omitempty" gorm:"column:shipping_fee"`
	ServiceFee          float64                `json:"service_fee,omitempty" gorm:"column:service_fee"`
	Paid                float64                `json:"paid,omitempty" gorm:"column:paid"`
	TotalBeforeTax      float64                `json:"total_before_tax,omitempty" gorm:"column:total_before_tax"`
	TotalBeforeDisc     float64                `json:"total_before_disc,omitempty" gorm:"column:total_before_disc"`
	Status              string                 `json:"status,omitempty" gorm:"column:status"`
	StockStatus         string                 `json:"stock_status,omitempty" gorm:"default:'pending';column:stock_status"`
	SalesDate           time.Time              `json:"sales_date,omitempty" gorm:"column:sales_date"`
	DueDate             time.Time              `json:"due_date,omitempty" gorm:"column:due_date"`
	PaymentTerms        string                 `json:"payment_terms,omitempty" gorm:"column:payment_terms"`
	PaymentID           *string                `json:"payment_id,omitempty" gorm:"column:payment_id"`
	Payment             *PaymentModel          `gorm:"foreignKey:PaymentID;constraint:OnDelete:CASCADE" json:"payment,omitempty"`
	OfferID             *string                `json:"offer_id,omitempty" gorm:"column:offer_id"`
	Offer               *OfferModel            `gorm:"foreignKey:OfferID;constraint:OnDelete:CASCADE" json:"offer,omitempty"`
	MerchantID          *string                `json:"merchant_id,omitempty" gorm:"column:merchant_id"`
	Merchant            *MerchantModel         `gorm:"foreignKey:MerchantID;constraint:OnDelete:CASCADE" json:"merchant,omitempty"`
	CompanyID           *string                `json:"company_id,omitempty" gorm:"column:company_id"`
	Company             *CompanyModel          `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE" json:"company,omitempty"`
	ContactID           *string                `json:"contact_id,omitempty" gorm:"column:contact_id"`
	Contact             *ContactModel          `gorm:"foreignKey:ContactID;constraint:OnDelete:CASCADE" json:"contact,omitempty"`
	ContactData         string                 `json:"-" gorm:"type:json;column:contact_data"`
	DataContact         map[string]interface{} `json:"data_contact" gorm:"-"`
	PaymentType         string                 `json:"payment_type,omitempty" gorm:"column:payment_type"`
	PaymentProviderType PaymentProviderType    `json:"payment_provider_type,omitempty" gorm:"column:payment_provider_type"`
	Items               []POSSalesItemModel    `json:"items,omitempty" gorm:"foreignKey:SalesID;constraint:OnDelete:CASCADE"`
	SaleAccountID       *string                `json:"sale_account_id,omitempty" gorm:"column:sale_account_id"`
	SaleAccount         *AccountModel          `gorm:"foreignKey:SaleAccountID;constraint:OnDelete:CASCADE" json:"sale_account,omitempty"`
	AssetAccountID      *string                `json:"asset_account_id,omitempty" gorm:"column:asset_account_id"`
	AssetAccount        *AccountModel          `gorm:"foreignKey:AssetAccountID;constraint:OnDelete:CASCADE" json:"asset_account,omitempty"`
	Tax                 float64                `json:"tax"`
	TaxType             string                 `json:"tax_type" gorm:"type:varchar"`
	TaxAmount           float64                `json:"tax_amount"`
}

type POSSalesItemModel struct {
	shared.BaseModel
	SalesID            string          `json:"sales_id,omitempty"`
	Sales              POSModel        `gorm:"foreignKey:SalesID;constraint:OnDelete:CASCADE" json:"-"`
	Description        string          `json:"description,omitempty"`
	Quantity           float64         `json:"quantity,omitempty"`
	UnitPrice          float64         `json:"unit_price,omitempty"`
	Total              float64         `json:"total,omitempty"`
	DiscountPercent    float64         `json:"discount_percent,omitempty"`
	DiscountAmount     float64         `json:"discount_amount,omitempty"`
	SubtotalBeforeDisc float64         `json:"subtotal_before_disc,omitempty"`
	ProductID          *string         `json:"product_id,omitempty"`
	Product            *ProductModel   `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE" json:"product,omitempty"`
	VariantID          *string         `json:"variant_id,omitempty"`
	Variant            *VariantModel   `gorm:"foreignKey:VariantID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"variant,omitempty"`
	WarehouseID        *string         `json:"warehouse_id,omitempty"`
	Warehouse          *WarehouseModel `gorm:"foreignKey:WarehouseID;constraint:OnDelete:CASCADE" json:"warehouse,omitempty"`
}

func (s *POSModel) TableName() string {
	return "pos_sales"
}

func (s *POSModel) AfterFind(tx *gorm.DB) (err error) {
	err = json.Unmarshal([]byte(s.ContactData), &s.DataContact)
	if err != nil {
		return err
	}
	return
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
