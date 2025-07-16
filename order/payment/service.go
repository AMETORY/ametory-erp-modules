package payment

import (
	"fmt"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/order/payment/payment_provider"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"gorm.io/gorm"
)

// PaymentService provides various methods to interact with payment providers
// and manage payment-related data.
type PaymentService struct {
	ctx             *context.ERPContext
	db              *gorm.DB
	PaymentProvider map[string]payment_provider.PaymentProvider
	activeProvider  string
}

// NewPaymentService creates a new instance of PaymentService.
func NewPaymentService(db *gorm.DB, ctx *context.ERPContext) *PaymentService {
	return &PaymentService{
		ctx:             ctx,
		db:              ctx.DB,
		PaymentProvider: make(map[string]payment_provider.PaymentProvider, 0),
	}
}

// AddPaymentProvider adds a new payment provider to the service and sets it
// as the active provider.
func (s *PaymentService) AddPaymentProvider(providerName string, paymentProvider payment_provider.PaymentProvider) {
	s.PaymentProvider[providerName] = paymentProvider
	s.SetActivePaymentProvider(providerName)
}

// SetActivePaymentProvider sets the active payment provider.
func (s *PaymentService) SetActivePaymentProvider(providerName string) {
	s.activeProvider = providerName
}

// Migrate applies database migrations for the PaymentModel.
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&models.PaymentModel{})
}

// CreatePaymentLink creates a payment link using the active payment provider.
func (s *PaymentService) CreatePaymentLink(data interface{}) (interface{}, error) {
	fmt.Println("PROVIDER", s.activeProvider)
	resp, err := s.PaymentProvider[s.activeProvider].CreatePaymentLink(data)
	return resp, err
}

// CreatePaymentVA creates a virtual account payment using the active payment provider.
func (s *PaymentService) CreatePaymentVA(data interface{}) (interface{}, error) {
	fmt.Println("PROVIDER", s.activeProvider)
	resp, err := s.PaymentProvider[s.activeProvider].CreatePaymentVA(data)
	return resp, err
}

// CreatePaymentEWallet creates an e-wallet payment using the active payment provider.
func (s *PaymentService) CreatePaymentEWallet(data interface{}) (interface{}, error) {
	fmt.Println("PROVIDER", s.activeProvider)
	resp, err := s.PaymentProvider[s.activeProvider].CreatePaymentEWallet(data)
	return resp, err
}

// DetailPayment retrieves payment details from the active payment provider.
func (s *PaymentService) DetailPayment(data ...interface{}) (interface{}, error) {
	resp, err := s.PaymentProvider[s.activeProvider].DetailPayment(data)
	return resp, err
}

// DetailPaymentVA retrieves virtual account payment details from the active payment provider.
func (s *PaymentService) DetailPaymentVA(data ...interface{}) (interface{}, error) {
	resp, err := s.PaymentProvider[s.activeProvider].DetailPaymentVA(data)
	return resp, err
}

// DetailPaymentEWallet retrieves e-wallet payment details from the active payment provider.
func (s *PaymentService) DetailPaymentEWallet(data ...interface{}) (interface{}, error) {
	resp, err := s.PaymentProvider[s.activeProvider].DetailPaymentEWallet(data)
	return resp, err
}

// CreatePayment creates a new payment record in the database.
func (s *PaymentService) CreatePayment(data *models.PaymentModel) error {
	data.PaymentData = "{}"
	return s.ctx.DB.Create(data).Error
}

// GetPaymentByCode retrieves a payment record by its code.
func (s *PaymentService) GetPaymentByCode(code string) (*models.PaymentModel, error) {
	data := models.PaymentModel{}
	err := s.ctx.DB.Where("code = ?", code).First(&data).Error
	return &data, err
}

// GetPaymentBankCode retrieves a map of bank codes.
func (s *PaymentService) GetPaymentBankCode() map[string]string {
	return models.BankCodes
}
