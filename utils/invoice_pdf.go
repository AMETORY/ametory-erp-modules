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
