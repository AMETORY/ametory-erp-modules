package payment_provider

// PaymentProvider is an interface for payment gateway providers
type PaymentProvider interface {
	CreatePaymentLink(data interface{}) (interface{}, error)
	DetailPayment(...interface{}) (interface{}, error)
}
