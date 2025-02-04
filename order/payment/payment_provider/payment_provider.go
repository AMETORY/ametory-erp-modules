package payment_provider

// PaymentProvider is an interface for payment gateway providers
type PaymentProvider interface {
	CreatePaymentLink(data interface{}) (interface{}, error)
	CreatePaymentVA(dataPayment interface{}) (interface{}, error)
	CreatePaymentEWallet(dataPayment interface{}) (interface{}, error)
	DetailPayment(...interface{}) (interface{}, error)
	DetailPaymentVA(...interface{}) (interface{}, error)
	DetailPaymentEWallet(...interface{}) (interface{}, error)
}
