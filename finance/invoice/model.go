package invoice

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/company"
	"github.com/AMETORY/ametory-erp-modules/contact"
	"github.com/AMETORY/ametory-erp-modules/utils"
)

type InvoiceModel struct {
	utils.BaseModel
	Code            string               `json:"code"`
	Description     string               `json:"description"`
	Notes           string               `json:"notes"`
	Total           float64              `json:"total"`
	Subtotal        float64              `json:"subtotal"`
	TotalBeforeTax  float64              `json:"total_before_tax"`
	TotalBeforeDisc float64              `json:"total_before_disc"`
	Status          string               `json:"status"`
	InvoiceDate     time.Time            `json:"invoice_date"`
	DueDate         time.Time            `json:"due_date"`
	PaymentTerms    string               `json:"payment_terms"`
	CompanyID       string               `json:"company_id"`
	Company         company.CompanyModel `gorm:"foreignKey:CompanyID"`
	ContactID       string               `json:"contact_id"`
	Contact         contact.ContactModel `gorm:"foreignKey:ContactID"`
	ContactData     string               `gorm:"type:json" json:"contact_data"`
}

type InvoiceItemModel struct {
	utils.BaseModel
	InvoiceID          string       `json:"invoice_id"`
	Invoice            InvoiceModel `gorm:"foreignKey:InvoiceID"`
	Description        string       `json:"description"`
	Quantity           float64      `json:"quantity"`
	UnitPrice          float64      `json:"unit_price"`
	Total              float64      `json:"total"`
	DiscountPercent    float64      `json:"discount_percent"`
	DiscountAmount     float64      `json:"discount_amount"`
	SubtotalBeforeDisc float64      `json:"subtotal_before_disc"`
}
