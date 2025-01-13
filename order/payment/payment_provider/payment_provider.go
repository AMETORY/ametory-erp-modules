package payment_provider

import (
	"errors"
)

// PaymentProvider is an interface for payment gateway providers
type PaymentProvider interface {
	// ProcessPayment processes a payment transaction
	ProcessPayment(amount float64, paymentMethod string) error

	// GetPaymentMethods returns a list of supported payment methods
	GetPaymentMethods() []string
}

// PaymentProviderFactory is a factory for creating payment providers
type PaymentProviderFactory struct{}

// NewPaymentProviderFactory returns a new payment provider factory
func NewPaymentProviderFactory() *PaymentProviderFactory {
	return &PaymentProviderFactory{}
}

// CreatePaymentProvider creates a payment provider based on the given payment method
func (f *PaymentProviderFactory) CreatePaymentProvider(paymentMethod string) (PaymentProvider, error) {
	switch paymentMethod {
	case "credit_card":
		return &CreditCardProvider{}, nil
	case "paypal":
		return &PayPalProvider{}, nil
	case "bank_transfer":
		return &BankTransferProvider{}, nil
	default:
		return nil, errors.New("unsupported payment method")
	}
}

// CreditCardProvider is a payment provider for credit card transactions
type CreditCardProvider struct{}

// ProcessPayment processes a credit card payment transaction
func (c *CreditCardProvider) ProcessPayment(amount float64, paymentMethod string) error {
	// Implement credit card payment processing logic here
	return nil
}

// GetPaymentMethods returns a list of supported payment methods for credit card
func (c *CreditCardProvider) GetPaymentMethods() []string {
	return []string{"visa", "mastercard", "amex"}
}

// PayPalProvider is a payment provider for PayPal transactions
type PayPalProvider struct{}

// ProcessPayment processes a PayPal payment transaction
func (p *PayPalProvider) ProcessPayment(amount float64, paymentMethod string) error {
	// Implement PayPal payment processing logic here
	return nil
}

// GetPaymentMethods returns a list of supported payment methods for PayPal
func (p *PayPalProvider) GetPaymentMethods() []string {
	return []string{"paypal"}
}

// BankTransferProvider is a payment provider for bank transfer transactions
type BankTransferProvider struct{}

// ProcessPayment processes a bank transfer payment transaction
func (b *BankTransferProvider) ProcessPayment(amount float64, paymentMethod string) error {
	// Implement bank transfer payment processing logic here
	return nil
}

// GetPaymentMethods returns a list of supported payment methods for bank transfer
func (b *BankTransferProvider) GetPaymentMethods() []string {
	return []string{"bca", "mandiri", "bri"}
}
