package pos

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/company"
	"github.com/AMETORY/ametory-erp-modules/contact"
	"github.com/AMETORY/ametory-erp-modules/finance/account"
	"github.com/AMETORY/ametory-erp-modules/inventory/product"
	"github.com/AMETORY/ametory-erp-modules/inventory/warehouse"
	"github.com/AMETORY/ametory-erp-modules/utils"
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
	utils.BaseModel
	SalesNumber         string                `json:"sales_number"`
	Code                string                `json:"code"`
	Description         string                `json:"description"`
	Notes               string                `json:"notes"`
	Total               float64               `json:"total"`
	Subtotal            float64               `json:"subtotal"`
	Paid                float64               `json:"paid"`
	TotalBeforeTax      float64               `json:"total_before_tax"`
	TotalBeforeDisc     float64               `json:"total_before_disc"`
	Status              string                `json:"status"`
	StockStatus         string                `json:"stock_status" gorm:"default:'pending'"`
	SalesDate           time.Time             `json:"sales_date"`
	DueDate             time.Time             `json:"due_date"`
	PaymentTerms        string                `json:"payment_terms"`
	MerchantID          *string               `json:"merchant_id"`
	Merchant            *MerchantModel        `gorm:"foreignKey:MerchantID"`
	CompanyID           string                `json:"company_id"`
	Company             company.CompanyModel  `gorm:"foreignKey:CompanyID"`
	ContactID           string                `json:"contact_id"`
	Contact             contact.ContactModel  `gorm:"foreignKey:ContactID"`
	ContactData         string                `gorm:"type:json" json:"contact_data"`
	PaymentType         string                `json:"payment_type"`
	PaymentProviderType PaymentProviderType   `json:"payment_provider_type"`
	Items               []POSSalesItemModel   `gorm:"foreignKey:SalesID" json:"items"`
	SaleAccountID       *string               `json:"sale_account_id"`
	SaleAccount         *account.AccountModel `gorm:"foreignKey:SaleAccountID"`
	AssetAccountID      *string               `json:"asset_account_id"`
	AssetAccount        *account.AccountModel `gorm:"foreignKey:AssetAccountID"`
}

type POSSalesItemModel struct {
	utils.BaseModel
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
	return db.AutoMigrate(&POSModel{}, &POSSalesItemModel{}, &MerchantModel{})
}
