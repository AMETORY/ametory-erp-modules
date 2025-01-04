package sales

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/company"
	"github.com/AMETORY/ametory-erp-modules/contact"
	"github.com/AMETORY/ametory-erp-modules/utils"
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
	utils.BaseModel
	SalesNumber     string               `json:"sales_number"`
	Code            string               `json:"code"`
	Description     string               `json:"description"`
	Notes           string               `json:"notes"`
	Total           float64              `json:"total"`
	Subtotal        float64              `json:"subtotal"`
	TotalBeforeTax  float64              `json:"total_before_tax"`
	TotalBeforeDisc float64              `json:"total_before_disc"`
	Status          string               `json:"status"`
	SalesDate       time.Time            `json:"sales_date"`
	DueDate         time.Time            `json:"due_date"`
	PaymentTerms    string               `json:"payment_terms"`
	CompanyID       string               `json:"company_id"`
	Company         company.CompanyModel `gorm:"foreignKey:CompanyID"`
	ContactID       string               `json:"contact_id"`
	Contact         contact.ContactModel `gorm:"foreignKey:ContactID"`
	ContactData     string               `gorm:"type:json" json:"contact_data"`
	Type            SalesType            `json:"type"`
}

type SalesItemModel struct {
	utils.BaseModel
	SalesID            string     `json:"sales_id"`
	Sales              SalesModel `gorm:"foreignKey:SalesID"`
	Description        string     `json:"description"`
	Quantity           float64    `json:"quantity"`
	UnitPrice          float64    `json:"unit_price"`
	Total              float64    `json:"total"`
	DiscountPercent    float64    `json:"discount_percent"`
	DiscountAmount     float64    `json:"discount_amount"`
	SubtotalBeforeDisc float64    `json:"subtotal_before_disc"`
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
	return "sales_item"
}

func (s *SalesItemModel) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&SalesModel{}, &SalesItemModel{})
}
