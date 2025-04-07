package models

import (
	"encoding/json"
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SalesType string
type SalesDocType string

const (
	ONLINE         SalesType = "ONLINE"
	OFFLINE        SalesType = "OFFLINE"
	DIRECT_SELLING SalesType = "DIRECT_SELLING"
)

const (
	INVOICE     SalesDocType = "INVOICE"
	SALES_ORDER SalesDocType = "SALES_ORDER"
	SALES_QUOTE SalesDocType = "SALES_QUOTE"
	DELIVERY    SalesDocType = "DELIVERY"
)

type SalesModel struct {
	shared.BaseModel
	SalesNumber           string              `json:"sales_number"`
	Code                  string              `json:"code"`
	Description           string              `json:"description"`
	Notes                 string              `json:"notes"`
	Total                 float64             `json:"total"`
	Subtotal              float64             `json:"subtotal"`
	Paid                  float64             `json:"paid"`
	TotalBeforeTax        float64             `json:"total_before_tax"`
	TotalBeforeDisc       float64             `json:"total_before_disc"`
	TotalTax              float64             `json:"total_tax"`
	TotalDiscount         float64             `json:"total_discount"`
	Status                string              `json:"status"`
	StockStatus           string              `json:"stock_status" gorm:"default:'pending'"`
	SalesDate             time.Time           `json:"sales_date"`
	DueDate               *time.Time          `json:"due_date"`
	DiscountDueDate       *time.Time          `json:"discount_due_date"`
	PaymentDiscountAmount float64             `json:"payment_discount_amount"`
	PaymentTerms          string              `json:"payment_terms"`
	PaymentTermsCode      string              `json:"payment_terms_code"`
	TermCondition         string              `json:"term_condition"`
	CompanyID             *string             `json:"company_id"`
	Company               *CompanyModel       `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE" json:"company"`
	UserID                *string             `gorm:"size:36" json:"-"`
	User                  *UserModel          `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	ContactID             *string             `json:"contact_id"`
	Contact               *ContactModel       `gorm:"foreignKey:ContactID;constraint:OnDelete:CASCADE" json:"contact"`
	ContactData           string              `gorm:"type:json" json:"contact_data"`
	DeliveryID            *string             `json:"delivery_id"`
	Delivery              *ContactModel       `gorm:"foreignKey:DeliveryID;constraint:OnDelete:CASCADE" json:"delivery"`
	DeliveryData          string              `gorm:"type:json" json:"delivery_data"`
	Type                  SalesType           `json:"type"`
	DocumentType          SalesDocType        `json:"document_type"`
	Items                 []SalesItemModel    `gorm:"foreignKey:SalesID;constraint:OnDelete:CASCADE" json:"items"`
	WithdrawalID          *string             `json:"withdrawal_id,omitempty" gorm:"column:withdrawal_id"`
	Withdrawal            *WithdrawalModel    `gorm:"foreignKey:WithdrawalID;constraint:OnDelete:CASCADE" json:"withdrawal,omitempty"`
	PublishedAt           *time.Time          `json:"published_at"`
	PublishedByID         *string             `json:"published_by_id,omitempty" gorm:"column:published_by_id"`
	PublishedBy           *UserModel          `gorm:"foreignKey:PublishedByID;constraint:OnDelete:CASCADE" json:"published_by,omitempty"`
	Taxes                 []*TaxModel         `gorm:"many2many:sales_taxes;constraint:OnDelete:CASCADE;" json:"taxes"`
	IsCompound            bool                `json:"is_compound"`
	TaxBreakdown          string              `gorm:"type:json" json:"tax_breakdown"`
	RefID                 *string             `json:"ref_id,omitempty"`
	RefType               *string             `gorm:"ref_type" json:"ref_type,omitempty"`
	SecondaryRefID        *string             `json:"secondary_ref_id,omitempty"`
	SecondaryRefType      *string             `gorm:"secondary_ref_type" json:"secondary_ref_type,omitempty"`
	ContactDataParsed     map[string]any      `json:"contact_data_parsed" gorm:"-"`
	DeliveryDataParsed    map[string]any      `json:"delivery_data_parsed" gorm:"-"`
	TaxBreakdownParsed    map[string]any      `json:"tax_breakdown_parsed" gorm:"-"`
	SalesRef              *SalesModel         `json:"sales_ref" gorm:"-"`
	SecondarySalesRef     *SalesModel         `json:"secondary_sales_ref" gorm:"-"`
	PaymentAccountID      *string             `json:"payment_account_id,omitempty"`
	PaymentAccount        *AccountModel       `json:"payment_account,omitempty" gorm:"foreignKey:PaymentAccountID;constraint:OnDelete:CASCADE"`
	SalesPayments         []SalesPaymentModel `gorm:"foreignKey:SalesID;constraint:OnDelete:CASCADE" json:"sales_payments"`
}

func (s *SalesModel) AfterFind(tx *gorm.DB) (err error) {
	var contactData map[string]any
	if err = json.Unmarshal([]byte(s.ContactData), &contactData); err != nil {
		return err
	}
	var deliveryDataParsed map[string]any
	if err = json.Unmarshal([]byte(s.DeliveryData), &deliveryDataParsed); err != nil {
		return err
	}
	var taxBreakdownParsed map[string]any
	if err = json.Unmarshal([]byte(s.TaxBreakdown), &taxBreakdownParsed); err != nil {
		return err
	}
	if s.DeliveryID != nil {
		var delivery ContactModel
		if err = tx.Model(&ContactModel{}).Where("id = ?", s.DeliveryID).First(&delivery).Error; err != nil {
			return err
		}
		s.Delivery = &delivery
	}
	s.ContactDataParsed = contactData
	s.DeliveryDataParsed = deliveryDataParsed
	s.TaxBreakdownParsed = taxBreakdownParsed
	return nil
}

type SalesItemModel struct {
	shared.BaseModel
	SalesID            *string         `json:"sales_id,omitempty"`
	Sales              *SalesModel     `gorm:"foreignKey:SalesID;constraint:OnDelete:CASCADE" json:"sales,omitempty"`
	Description        string          `json:"description,omitempty"`
	Notes              string          `json:"notes,omitempty"`
	Quantity           float64         `json:"quantity,omitempty"`
	BasePrice          float64         `json:"base_price,omitempty"`
	UnitPrice          float64         `json:"unit_price,omitempty"`
	Total              float64         `json:"total,omitempty"`
	SubTotal           float64         `json:"sub_total,omitempty"`
	DiscountPercent    float64         `json:"discount_percent,omitempty"`
	DiscountAmount     float64         `json:"discount_amount,omitempty"`
	SubtotalBeforeDisc float64         `json:"subtotal_before_disc,omitempty"`
	ProductID          *string         `json:"product_id,omitempty"`
	Product            *ProductModel   `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE" json:"product,omitempty"`
	VariantID          *string         `json:"variant_id,omitempty"`
	Variant            *VariantModel   `gorm:"foreignKey:VariantID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"variant,omitempty"`
	WarehouseID        *string         `json:"warehouse_id,omitempty"`
	Warehouse          *WarehouseModel `gorm:"foreignKey:WarehouseID;constraint:OnDelete:CASCADE" json:"warehouse,omitempty"`
	SaleAccountID      *string         `json:"sale_account_id,omitempty"`
	SaleAccount        *AccountModel   `gorm:"foreignKey:SaleAccountID;constraint:OnDelete:CASCADE" json:"sale_account,omitempty"`
	AssetAccountID     *string         `json:"asset_account_id,omitempty"`
	AssetAccount       *AccountModel   `gorm:"foreignKey:AssetAccountID;constraint:OnDelete:CASCADE" json:"asset_account,omitempty"`
	TaxID              *string         `json:"tax_id,omitempty"`
	Tax                *TaxModel       `gorm:"foreignKey:TaxID;constraint:Restrict:SET NULL" json:"tax,omitempty"`
	TotalTax           float64         `json:"total_tax,omitempty"`
	UnitID             *string         `json:"unit_id,omitempty"`
	Unit               *UnitModel      `gorm:"foreignKey:UnitID;constraint:OnDelete:CASCADE" json:"unit,omitempty"`
	UnitValue          float64         `json:"unit_value,omitempty" gorm:"default:1"`
	IsCost             bool            `json:"is_cost,omitempty" gorm:"default:false"`
}

func (s *SalesModel) TableName() string {
	return "sales"
}

func (s *SalesModel) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}

func (s *SalesItemModel) TableName() string {
	return "sales_items"
}

func (s *SalesItemModel) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}

type SalesPaymentModel struct {
	shared.BaseModel
	PaymentDate        time.Time     `json:"payment_date"`
	SalesID            *string       `json:"sales_id"`
	Sales              *SalesModel   `gorm:"foreignKey:SalesID;constraint:OnDelete:CASCADE" json:"sales"`
	Amount             float64       `json:"amount"`
	PaymentDiscount    float64       `json:"payment_discount"`
	Notes              string        `json:"notes"`
	CompanyID          *string       `json:"company_id"`
	Company            *CompanyModel `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE" json:"company"`
	UserID             *string       `gorm:"size:36" json:"-"`
	User               *UserModel    `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	AssetAccountID     *string       `json:"asset_account_id"`
	AssetAccount       *AccountModel `gorm:"foreignKey:AssetAccountID;constraint:OnDelete:CASCADE" json:"asset_account"`
	IsRefund           bool          `json:"is_refund"`
	PaymentMethod      string        `gorm:"default:CASH" json:"payment_method"`
	PaymentMethodNotes string        `json:"payment_method_notes"`
}

func (s *SalesPaymentModel) TableName() string {
	return "sales_payments"
}

func (s *SalesPaymentModel) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}
