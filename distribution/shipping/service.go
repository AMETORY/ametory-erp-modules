package shipping

import (
	"errors"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/distribution/shipping/provider"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/shared/objects"
	"gorm.io/gorm"
)

type ShippingService struct {
	db       *gorm.DB
	ctx      *context.ERPContext
	provider provider.ShippingProvider
}

func NewShippingService(db *gorm.DB, ctx *context.ERPContext) *ShippingService {
	return &ShippingService{db: db, ctx: ctx}
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&models.ShippingModel{})
}
func (s *ShippingService) SetProvider(provider provider.ShippingProvider) {
	s.provider = provider
}
func (s *ShippingService) CreateShipping(orderID string, method, destination string) (*models.ShippingModel, error) {
	if s.provider == nil {
		return nil, errors.New("shipping provider not set")
	}
	// Buat pengiriman menggunakan provider
	shipment, err := s.provider.CreateShipment(orderID, destination)
	if err != nil {
		return nil, err
	}

	// Simpan data pengiriman ke database
	shipping := models.ShippingModel{
		OrderID:    orderID,
		Method:     method,
		TrackingID: shipment.TrackingID,
		Status:     shipment.Status,
	}
	if err := s.db.Create(&shipping).Error; err != nil {
		return nil, err
	}

	return &shipping, nil
}

func (s *ShippingService) GetRates(origin, destination string, weight float64) ([]objects.Rate, error) {
	if s.provider == nil {
		return nil, errors.New("shipping provider not set")
	}
	return s.provider.GetRates(origin, destination, weight)
}

func (s *ShippingService) GetExpressMotorRates(origin, destination objects.LocationPrecise) ([]objects.Rate, error) {
	if s.provider == nil {
		return nil, errors.New("shipping provider not set")
	}
	return s.provider.GetExpressMotorRates(origin, destination)
}

func (s *ShippingService) TrackShipment(trackingID string) (*objects.TrackingStatus, error) {
	if s.provider == nil {
		return nil, errors.New("shipping provider not set")
	}

	status, err := s.provider.TrackShipment(trackingID)
	if err != nil {
		return nil, err
	}

	return &status, nil
}
