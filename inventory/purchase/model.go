package purchase

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

type PurchaseType string

const (
	PURCHASE    PurchaseType = "PURCHASE"
	PROCUREMENT PurchaseType = "PROCUREMENT"
	ECOMMERCE   PurchaseType = "ECOMMERCE"
)

type PurchaseOrderModel struct {
	utils.BaseModel
	PurchaseNumber  string                   `json:"purchase_number"`
	Code            string                   `json:"code"`
	Description     string                   `json:"description"`
	Notes           string                   `json:"notes"`
	Total           float64                  `json:"total"`
	Paid            float64                  `json:"paid"`
	Subtotal        float64                  `json:"subtotal"`
	TotalBeforeTax  float64                  `json:"total_before_tax"`
	TotalBeforeDisc float64                  `json:"total_before_disc"`
	Status          string                   `json:"status"`
	StockStatus     string                   `json:"stock_status" gorm:"default:'pending'"`
	PurchaseDate    time.Time                `json:"purchase_date"`
	DueDate         time.Time                `json:"due_date"`
	PaymentTerms    string                   `json:"payment_terms"`
	CompanyID       string                   `json:"company_id"`
	Company         company.CompanyModel     `gorm:"foreignKey:CompanyID"`
	ContactID       string                   `json:"contact_id"`
	Contact         contact.ContactModel     `gorm:"foreignKey:ContactID"`
	ContactData     string                   `gorm:"type:json" json:"contact_data"`
	Type            PurchaseType             `json:"type"`
	Items           []PurchaseOrderItemModel `gorm:"foreignKey:PurchaseID" json:"items"`
}

type PurchaseOrderItemModel struct {
	utils.BaseModel
	PurchaseID         string                    `json:"purchase_id"`
	Purchase           PurchaseOrderModel        `gorm:"foreignKey:PurchaseID"`
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
	PurchaseAccountID  *string                   `json:"purchase_account_id"`
	PurchaseAccount    *account.AccountModel     `gorm:"foreignKey:PurchaseAccountID"`
	AssetAccountID     *string                   `json:"asset_account_id"`
	AssetAccount       *account.AccountModel     `gorm:"foreignKey:AssetAccountID"`
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

func (s *PurchaseOrderItemModel) TableName() string {
	return "purchase_order_items"
}

func (s *PurchaseOrderItemModel) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&PurchaseOrderModel{}, &PurchaseOrderItemModel{})
}
