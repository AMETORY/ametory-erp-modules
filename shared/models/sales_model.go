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
	Status          string           `json:"status"`
	StockStatus     string           `json:"stock_status" gorm:"default:'pending'"`
	SalesDate       time.Time        `json:"sales_date"`
	DueDate         time.Time        `json:"due_date"`
	PaymentTerms    string           `json:"payment_terms"`
	CompanyID       *string          `json:"company_id"`
	Company         *CompanyModel    `gorm:"foreignKey:CompanyID" json:"company"`
	ContactID       string           `json:"contact_id"`
	Contact         ContactModel     `gorm:"foreignKey:ContactID" json:"contact"`
	ContactData     string           `gorm:"type:json" json:"contact_data"`
	Type            SalesType        `json:"type"`
	Items           []SalesItemModel `gorm:"foreignKey:SalesID" json:"items"`
}

type SalesItemModel struct {
	shared.BaseModel
	SalesID            string          `json:"sales_id"`
	Sales              SalesModel      `gorm:"foreignKey:SalesID"`
	Description        string          `json:"description"`
	Quantity           float64         `json:"quantity"`
	UnitPrice          float64         `json:"unit_price"`
	Total              float64         `json:"total"`
	DiscountPercent    float64         `json:"discount_percent"`
	DiscountAmount     float64         `json:"discount_amount"`
	SubtotalBeforeDisc float64         `json:"subtotal_before_disc"`
	ProductID          *string         `json:"product_id"`
	Product            *ProductModel   `gorm:"foreignKey:ProductID"`
	WarehouseID        *string         `json:"warehouse_id"`
	Warehouse          *WarehouseModel `gorm:"foreignKey:WarehouseID"`
	SaleAccountID      *string         `json:"sale_account_id"`
	SaleAccount        *AccountModel   `gorm:"foreignKey:SaleAccountID"`
	AssetAccountID     *string         `json:"asset_account_id"`
	AssetAccount       *AccountModel   `gorm:"foreignKey:AssetAccountID"`
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
