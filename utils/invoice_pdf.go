package utils

type InvoicePDF struct {
	Company         InvoicePDFContact
	Number          string
	Date            string
	DueDate         string
	Items           []InvoicePDFItem
	SubTotal        string
	TotalDiscount   string
	AfterDiscount   string
	TotalTax        string
	GrandTotal      string
	InvoicePayments []InvoicePDFPayment
	Balance         string
	Paid            string
	BilledTo        InvoicePDFContact
	ShippedTo       InvoicePDFContact
	TermCondition   string
	PaymentTerms    string
	ShowCompany     bool
	ShowShipped     bool
}

type ReceiptData struct {
	Items           []ReceiptItem `json:"items"`
	SubTotalPrice   string        `json:"sub_total_price"`
	DiscountAmount  string        `json:"discount_amount"`
	TotalPrice      string        `json:"total_price"`
	CashierName     string        `json:"cashier_name"`
	Code            string        `json:"code"`
	Date            string        `json:"date"`
	CustomerName    string        `json:"customer_name"`
	MerchantName    string        `json:"merchant_name"`
	MerchantAddress string        `json:"merchant_address"`
}

type ReceiptItem struct {
	Description     string `json:"description"`
	Quantity        string `json:"quantity"`
	Price           string `json:"price"`
	Total           string `json:"total"`
	DiscountPercent string `json:"discount_percent"`
	Notes           string `json:"notes"`
}

type InvoicePDFItem struct {
	No                 int
	Description        string
	Notes              string
	Quantity           string
	UnitPrice          string
	UnitName           string
	Total              string
	SubTotal           string
	SubtotalBeforeDisc string
	TotalDiscount      string
	DiscountPercent    string
	TaxAmount          string
	TaxPercent         string
	TaxName            string
}

type InvoicePDFPayment struct {
	Date               string
	Description        string
	PaymentMethod      string
	Amount             string
	PaymentDiscount    string
	PaymentMethodNotes string
}

type InvoicePDFContact struct {
	Name    string
	Address string
	Phone   string
	Email   string
}
