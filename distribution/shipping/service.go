package shipping

import (
	"errors"

	"github.com/AMETORY/ametory-erp-modules/distribution/shipping/provider"
	"gorm.io/gorm"
)

type ShippingService struct {
	db       *gorm.DB
	provider provider.ShippingProvider
}

func NewShippingService(db *gorm.DB, provider provider.ShippingProvider) *ShippingService {
	return &ShippingService{db: db, provider: provider}
}
func (s *ShippingService) SetProvider(provider provider.ShippingProvider) {
	s.provider = provider
}
func (s *ShippingService) CreateShipping(orderID uint, method, destination string) (*ShippingModel, error) {
	if s.provider == nil {
		return nil, errors.New("shipping provider not set")
	}
	// Buat pengiriman menggunakan provider
	shipment, err := s.provider.CreateShipment(orderID, destination)
	if err != nil {
		return nil, err
	}

	// Simpan data pengiriman ke database
	shipping := ShippingModel{
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

func (s *ShippingService) GetRates(origin, destination string, weight float64) ([]provider.Rate, error) {
	if s.provider == nil {
		return nil, errors.New("shipping provider not set")
	}
	return s.provider.GetRates(origin, destination, weight)
}

func (s *ShippingService) TrackShipment(trackingID string) (*provider.TrackingStatus, error) {
	if s.provider == nil {
		return nil, errors.New("shipping provider not set")
	}

	status, err := s.provider.TrackShipment(trackingID)
	if err != nil {
		return nil, err
	}

	return &status, nil
}
