package models

import (
	"encoding/json"
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PurchaseType string
type PurchaseDocType string

const (
	BILL           PurchaseDocType = "BILL"
	PURCHASE_ORDER PurchaseDocType = "PURCHASE_ORDER"
)

const (
	PURCHASE           PurchaseType = "PURCHASE"
	PROCUREMENT        PurchaseType = "PROCUREMENT"
	PURCHASE_ECOMMERCE PurchaseType = "ECOMMERCE"
)

type PurchaseOrderModel struct {
	shared.BaseModel
	PurchaseNumber        string                   `json:"purchase_number,omitempty"`
	Code                  string                   `json:"code,omitempty"`
	Description           string                   `json:"description,omitempty"`
	Notes                 string                   `json:"notes,omitempty"`
	Total                 float64                  `json:"total,omitempty"`
	Paid                  float64                  `json:"paid,omitempty"`
	Subtotal              float64                  `json:"subtotal,omitempty"`
	TotalBeforeTax        float64                  `json:"total_before_tax,omitempty"`
	TotalBeforeDisc       float64                  `json:"total_before_disc,omitempty"`
	TotalTax              float64                  `json:"total_tax,omitempty"`
	TotalDiscount         float64                  `json:"total_discount,omitempty"`
	Status                string                   `json:"status,omitempty"`
	StockStatus           string                   `json:"stock_status,omitempty" gorm:"default:'pending'"`
	PurchaseDate          time.Time                `json:"purchase_date,omitempty"`
	DueDate               *time.Time               `json:"due_date,omitempty"`
	DiscountDueDate       *time.Time               `json:"discount_due_date,omitempty"`
	PaymentAccountID      *string                  `json:"payment_account_id,omitempty"`
	PaymentAccount        *AccountModel            `json:"payment_account,omitempty" gorm:"foreignKey:PaymentAccountID;constraint:OnDelete:CASCADE"`
	PaymentDiscountAmount float64                  `json:"payment_discount_amount,omitempty"`
	PaymentTerms          string                   `json:"payment_terms,omitempty"`
	PaymentTermsCode      string                   `json:"payment_terms_code,omitempty"`
	TermCondition         string                   `json:"term_condition,omitempty"`
	CompanyID             *string                  `json:"company_id,omitempty"`
	Company               *CompanyModel            `json:"company,omitempty" gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE"`
	UserID                *string                  `json:"user_id,omitempty" gorm:"size:36"`
	User                  *UserModel               `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	ContactID             *string                  `json:"contact_id,omitempty"`
	Contact               *ContactModel            `json:"contact,omitempty" gorm:"foreignKey:ContactID;constraint:OnDelete:CASCADE"`
	ContactData           string                   `json:"contact_data,omitempty" gorm:"type:json"`
	Type                  PurchaseType             `json:"type,omitempty"`
	DocumentType          PurchaseDocType          `json:"document_type,omitempty"`
	Items                 []PurchaseOrderItemModel `json:"items,omitempty" gorm:"foreignKey:PurchaseID;constraint:OnDelete:CASCADE"`
	PublishedAt           *time.Time               `json:"published_at,omitempty"`
	PublishedByID         *string                  `json:"published_by_id,omitempty" gorm:"column:published_by_id"`
	PublishedBy           *UserModel               `json:"published_by,omitempty" gorm:"foreignKey:PublishedByID;constraint:OnDelete:CASCADE"`
	RefID                 *string                  `json:"ref_id,omitempty"`
	RefType               *string                  `json:"ref_type,omitempty" gorm:"ref_type"`
	SecondaryRefID        *string                  `json:"secondary_ref_id,omitempty"`
	SecondaryRefType      *string                  `json:"secondary_ref_type,omitempty" gorm:"secondary_ref_type"`
	PurchaseRef           *PurchaseOrderModel      `json:"purchase_ref,omitempty" gorm:"-"`
	SecondaryPurchaseRef  *PurchaseOrderModel      `json:"secondary_purchase_ref,omitempty" gorm:"-"`
	Taxes                 []*TaxModel              `json:"taxes,omitempty" gorm:"many2many:sales_taxes;constraint:OnDelete:CASCADE;"`
	IsCompound            bool                     `json:"is_compound,omitempty"`
	TaxBreakdown          string                   `json:"tax_breakdown,omitempty" gorm:"type:json"`
	ContactDataParsed     map[string]any           `json:"contact_data_parsed" gorm:"-"`
	DeliveryDataParsed    map[string]any           `json:"delivery_data_parsed" gorm:"-"`
	TaxBreakdownParsed    map[string]any           `json:"tax_breakdown_parsed" gorm:"-"`
	PurchasePayments      []PurchasePaymentModel   `gorm:"foreignKey:PurchaseID;constraint:OnDelete:CASCADE" json:"purchase_payments"`
}

func (s *PurchaseOrderModel) TableName() string {
	return "purchase_orders"
}

func (s *PurchaseOrderModel) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}

func (s *PurchaseOrderModel) AfterFind(tx *gorm.DB) (err error) {
	var contactData map[string]any
	if err = json.Unmarshal([]byte(s.ContactData), &contactData); err != nil {
		return err
	}

	var taxBreakdownParsed map[string]any
	if err = json.Unmarshal([]byte(s.TaxBreakdown), &taxBreakdownParsed); err != nil {
		return err
	}

	s.ContactDataParsed = contactData
	s.TaxBreakdownParsed = taxBreakdownParsed
	return nil
}

type PurchaseOrderItemModel struct {
	shared.BaseModel
	PurchaseID         *string             `json:"purchase_id,omitempty"`
	Purchase           *PurchaseOrderModel `gorm:"foreignKey:PurchaseID;constraint:OnDelete:CASCADE"`
	Description        string              `json:"description,omitempty"`
	Notes              string              `gorm:"type:text" json:"notes"`
	Quantity           float64             `json:"quantity,omitempty"`
	UnitPrice          float64             `json:"unit_price,omitempty"`
	Total              float64             `json:"total,omitempty"`
	SubTotal           float64             `json:"sub_total,omitempty"`
	DiscountPercent    float64             `json:"discount_percent,omitempty"`
	DiscountAmount     float64             `json:"discount_amount,omitempty"`
	SubtotalBeforeDisc float64             `json:"subtotal_before_disc,omitempty"`
	ProductID          *string             `json:"product_id,omitempty"`
	Product            *ProductModel       `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE" json:"product,omitempty"`
	VariantID          *string             `json:"variant_id,omitempty" `
	Variant            *VariantModel       `gorm:"foreignKey:VariantID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"variant,omitempty"`
	WarehouseID        *string             `json:"warehouse_id,omitempty"`
	Warehouse          *WarehouseModel     `gorm:"foreignKey:WarehouseID;constraint:OnDelete:CASCADE" json:"warehouse,omitempty"`
	TaxID              *string             `json:"tax_id,omitempty"`
	Tax                *TaxModel           `gorm:"foreignKey:TaxID;constraint:Restrict:SET NULL" json:"tax,omitempty"`
	TotalTax           float64             `json:"total_tax,omitempty"`
	UnitID             *string             `json:"unit_id,omitempty"`
	Unit               *UnitModel          `gorm:"foreignKey:UnitID;constraint:OnDelete:CASCADE" json:"unit,omitempty"`
	UnitValue          float64             `json:"unit_value,omitempty" gorm:"default:1"`
	IsCost             bool                `json:"is_cost,omitempty" gorm:"default:false"`
}

func (s *PurchaseOrderItemModel) TableName() string {
	return "purchase_order_items"
}

func (s *PurchaseOrderItemModel) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return

}

type PurchasePaymentModel struct {
	shared.BaseModel
	PaymentDate        time.Time           `json:"payment_date"`
	PurchaseID         *string             `json:"purchase_id"`
	Purchase           *PurchaseOrderModel `gorm:"foreignKey:PurchaseID;constraint:OnDelete:CASCADE" json:"sales"`
	Amount             float64             `json:"amount"`
	PaymentDiscount    float64             `json:"payment_discount"`
	Notes              string              `json:"notes"`
	CompanyID          *string             `json:"company_id"`
	Company            *CompanyModel       `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE" json:"company"`
	UserID             *string             `gorm:"size:36" json:"-"`
	User               *UserModel          `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	AssetAccountID     *string             `json:"asset_account_id"`
	AssetAccount       *AccountModel       `gorm:"foreignKey:AssetAccountID;constraint:OnDelete:CASCADE" json:"asset_account"`
	IsRefund           bool                `json:"is_refund"`
	PaymentMethod      string              `gorm:"default:CASH" json:"payment_method"`
	PaymentMethodNotes string              `json:"payment_method_notes"`
}

func (s *PurchasePaymentModel) TableName() string {
	return "purchase_payments"
}

func (s *PurchasePaymentModel) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}
