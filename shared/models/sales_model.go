package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SalesType string

const (
	INVOICE        SalesType = "INVOICE"
	POS            SalesType = "POS"
	DIRECT_SELLING SalesType = "DIRECT_SELLING"
	ECOMMERCE      SalesType = "ECOMMERCE"
)

type SalesModel struct {
	shared.BaseModel
	SalesNumber     string           `json:"sales_number"`
	Code            string           `json:"code"`
	Description     string           `json:"description"`
	Notes           string           `json:"notes"`
	Total           float64          `json:"total"`
	Subtotal        float64          `json:"subtotal"`
	Paid            float64          `json:"paid"`
	TotalBeforeTax  float64          `json:"total_before_tax"`
	TotalBeforeDisc float64          `json:"total_before_disc"`
	TotalTax        float64          `json:"total_tax"`
	Status          string           `json:"status"`
	StockStatus     string           `json:"stock_status" gorm:"default:'pending'"`
	SalesDate       time.Time        `json:"sales_date"`
	DueDate         time.Time        `json:"due_date"`
	PaymentTerms    string           `json:"payment_terms"`
	CompanyID       *string          `json:"company_id"`
	Company         *CompanyModel    `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE" json:"company"`
	UserID          *string          `gorm:"size:36" json:"-"`
	User            *UserModel       `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	ContactID       *string          `json:"contact_id"`
	Contact         *ContactModel    `gorm:"foreignKey:ContactID;constraint:OnDelete:CASCADE" json:"contact"`
	ContactData     string           `gorm:"type:json" json:"contact_data"`
	Type            SalesType        `json:"type"`
	Items           []SalesItemModel `gorm:"foreignKey:SalesID;constraint:OnDelete:CASCADE" json:"items"`
	WithdrawalID    *string          `json:"withdrawal_id,omitempty" gorm:"column:withdrawal_id"`
	Withdrawal      *WithdrawalModel `gorm:"foreignKey:WithdrawalID;constraint:OnDelete:CASCADE" json:"withdrawal,omitempty"`
	PublishedAt     *time.Time       `json:"published_at"`
	Taxes           []*TaxModel      `gorm:"many2many:sales_taxes;constraint:OnDelete:CASCADE;" json:"taxes"`
	IsCompound      bool             `json:"is_compound"`
	TaxBreakdown    string           `gorm:"type:json" json:"tax_breakdown"`
}

type SalesItemModel struct {
	shared.BaseModel
	SalesID            string          `json:"sales_id"`
	Sales              SalesModel      `gorm:"foreignKey:SalesID;constraint:OnDelete:CASCADE"`
	Description        string          `json:"description"`
	Quantity           float64         `json:"quantity"`
	UnitPrice          float64         `json:"unit_price"`
	Total              float64         `json:"total"`
	SubTotal           float64         `json:"sub_total"`
	DiscountPercent    float64         `json:"discount_percent"`
	DiscountAmount     float64         `json:"discount_amount"`
	SubtotalBeforeDisc float64         `json:"subtotal_before_disc"`
	ProductID          *string         `json:"product_id"`
	Product            *ProductModel   `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE"`
	VariantID          *string         `json:"variant_id,omitempty"`
	Variant            *VariantModel   `gorm:"foreignKey:VariantID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
	WarehouseID        *string         `json:"warehouse_id"`
	Warehouse          *WarehouseModel `gorm:"foreignKey:WarehouseID;constraint:OnDelete:CASCADE"`
	SaleAccountID      *string         `json:"sale_account_id"`
	SaleAccount        *AccountModel   `gorm:"foreignKey:SaleAccountID;constraint:OnDelete:CASCADE"`
	AssetAccountID     *string         `json:"asset_account_id"`
	AssetAccount       *AccountModel   `gorm:"foreignKey:AssetAccountID;constraint:OnDelete:CASCADE"`
	TaxID              *string         `json:"tax_id"`
	Tax                *TaxModel       `gorm:"foreignKey:TaxID;constraint:Restrict:SET NULL" json:"tax,omitempty"`
	TotalTax           float64         `json:"total_tax"`
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
