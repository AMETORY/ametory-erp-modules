package pos

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/company"
	"github.com/AMETORY/ametory-erp-modules/contact"
	"github.com/AMETORY/ametory-erp-modules/finance/account"
	"github.com/AMETORY/ametory-erp-modules/inventory/product"
	"github.com/AMETORY/ametory-erp-modules/inventory/warehouse"
	"github.com/AMETORY/ametory-erp-modules/order/merchant"
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
	SalesNumber         string                  `json:"sales_number" gorm:"column:sales_number"`
	Code                string                  `json:"code" gorm:"column:code"`
	Description         string                  `json:"description" gorm:"column:description"`
	Notes               string                  `json:"notes" gorm:"column:notes"`
	Total               float64                 `json:"total" gorm:"column:total"`
	Subtotal            float64                 `json:"subtotal" gorm:"column:subtotal"`
	Paid                float64                 `json:"paid" gorm:"column:paid"`
	TotalBeforeTax      float64                 `json:"total_before_tax" gorm:"column:total_before_tax"`
	TotalBeforeDisc     float64                 `json:"total_before_disc" gorm:"column:total_before_disc"`
	Status              string                  `json:"status" gorm:"column:status"`
	StockStatus         string                  `json:"stock_status" gorm:"default:'pending';column:stock_status"`
	SalesDate           time.Time               `json:"sales_date" gorm:"column:sales_date"`
	DueDate             time.Time               `json:"due_date" gorm:"column:due_date"`
	PaymentTerms        string                  `json:"payment_terms" gorm:"column:payment_terms"`
	MerchantID          *string                 `json:"merchant_id" gorm:"column:merchant_id"`
	Merchant            *merchant.MerchantModel `gorm:"foreignKey:MerchantID;column:merchant"`
	CompanyID           *string                 `json:"company_id" gorm:"column:company_id"`
	Company             *company.CompanyModel   `gorm:"foreignKey:CompanyID;column:company"`
	ContactID           string                  `json:"contact_id" gorm:"column:contact_id"`
	Contact             contact.ContactModel    `gorm:"foreignKey:ContactID;column:contact"`
	ContactData         string                  `json:"contact_data" gorm:"type:json;column:contact_data"`
	PaymentType         string                  `json:"payment_type" gorm:"column:payment_type"`
	PaymentProviderType PaymentProviderType     `json:"payment_provider_type" gorm:"column:payment_provider_type"`
	Items               []POSSalesItemModel     `json:"items" gorm:"foreignKey:SalesID;column:items"`
	SaleAccountID       *string                 `json:"sale_account_id" gorm:"column:sale_account_id"`
	SaleAccount         *account.AccountModel   `gorm:"foreignKey:SaleAccountID;column:sale_account"`
	AssetAccountID      *string                 `json:"asset_account_id" gorm:"column:asset_account_id"`
	AssetAccount        *account.AccountModel   `gorm:"foreignKey:AssetAccountID;column:asset_account"`
}
type POSSalesItemModel struct {
	shared.BaseModel
	SalesID            string                    `json:"sales_id"`
	Sales              POSModel                  `gorm:"foreignKey:SalesID"`
	Description        string                    `json:"description"`
	Quantity           float64                   `json:"quantity"`
	UnitPrice          float64                   `json:"unit_price"`
	Total              float64                   `json:"total"`
	DiscountPercent    float64                   `json:"discount_percent"`
	DiscountAmount     float64                   `json:"discount_amount"`
	SubtotalBeforeDisc float64                   `json:"subtotal_before_disc"`
	ProductID          *string                   `json:"product_id"`
	Product            *product.ProductModel     `gorm:"foreignKey:ProductID"`
	WarehouseID        *string                   `json:"warehouse_id"`
	Warehouse          *warehouse.WarehouseModel `gorm:"foreignKey:WarehouseID"`
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
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&POSModel{}, &POSSalesItemModel{})
}
