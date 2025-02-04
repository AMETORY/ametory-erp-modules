package payment

import (
	"fmt"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/order/payment/payment_provider"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"gorm.io/gorm"
)

type PaymentService struct {
	ctx             *context.ERPContext
	db              *gorm.DB
	PaymentProvider map[string]payment_provider.PaymentProvider
	activeProvider  string
}

func NewPaymentService(db *gorm.DB, ctx *context.ERPContext) *PaymentService {
	return &PaymentService{
		ctx:             ctx,
		db:              ctx.DB,
		PaymentProvider: make(map[string]payment_provider.PaymentProvider, 0),
	}
}

func (s *PaymentService) AddPaymentProvider(providerName string, paymentProvider payment_provider.PaymentProvider) {
	s.PaymentProvider[providerName] = paymentProvider
	s.SetActivePaymentProvider(providerName)
}

func (s *PaymentService) SetActivePaymentProvider(providerName string) {
	s.activeProvider = providerName
}
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&models.PaymentModel{})
}

func (s *PaymentService) CreatePaymentLink(data interface{}) (interface{}, error) {
	fmt.Println("PROVIDER", s.activeProvider)
	// fmt.Println("PROVIDER DATA", data)
	resp, err := s.PaymentProvider[s.activeProvider].CreatePaymentLink(data)
	return resp, err
}
func (s *PaymentService) CreatePaymentVA(data interface{}) (interface{}, error) {
	fmt.Println("PROVIDER", s.activeProvider)
	// fmt.Println("PROVIDER DATA", data)
	resp, err := s.PaymentProvider[s.activeProvider].CreatePaymentVA(data)
	return resp, err
}
func (s *PaymentService) CreatePaymentEWallet(data interface{}) (interface{}, error) {
	fmt.Println("PROVIDER", s.activeProvider)
	// fmt.Println("PROVIDER DATA", data)
	resp, err := s.PaymentProvider[s.activeProvider].CreatePaymentEWallet(data)
	return resp, err
}
func (s *PaymentService) DetailPayment(data ...interface{}) (interface{}, error) {
	resp, err := s.PaymentProvider[s.activeProvider].DetailPayment(data)
	return resp, err
}
func (s *PaymentService) DetailPaymentVA(data ...interface{}) (interface{}, error) {
	resp, err := s.PaymentProvider[s.activeProvider].DetailPaymentVA(data)
	return resp, err
}
func (s *PaymentService) DetailPaymentEWallet(data ...interface{}) (interface{}, error) {
	resp, err := s.PaymentProvider[s.activeProvider].DetailPaymentEWallet(data)
	return resp, err
}

func (s *PaymentService) CreatePayment(data *models.PaymentModel) error {
	data.PaymentData = "{}"
	return s.ctx.DB.Create(data).Error
}

func (s *PaymentService) GetPaymentByCode(code string) (*models.PaymentModel, error) {
	data := models.PaymentModel{}
	err := s.ctx.DB.Where("code = ?", code).First(&data).Error

	return &data, err
}

func (s *PaymentService) GetPaymentBankCode() map[string]string {
	return models.BankCodes
}
