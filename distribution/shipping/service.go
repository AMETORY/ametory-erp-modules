package shipping

import (
	"encoding/json"
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

func (s *ShippingService) CreateDraftShipment(data interface{}) (objects.Shipment, error) {
	return s.provider.CreateDraftShipment(data)
}

func (s *ShippingService) CreateShipment(data interface{}) (objects.Shipment, error) {
	return s.provider.CreateShipment(data)
}

func (s *ShippingService) GetShippingByOrderID(orderID string) (*models.ShippingModel, error) {
	var shipping models.ShippingModel
	if err := s.db.Where("order_id = ?", orderID).First(&shipping).Error; err != nil {
		return nil, err
	}
	return &shipping, nil
}
func (s *ShippingService) GetShippingByID(ID string) (*models.ShippingModel, error) {
	var shipping models.ShippingModel
	if err := s.db.Where("id = ?", ID).First(&shipping).Error; err != nil {
		return nil, err
	}
	return &shipping, nil
}

func (s *ShippingService) GetShippingByTrackingID(trackingID string) (*models.ShippingModel, error) {
	var shipping models.ShippingModel
	if err := s.db.Where("tracking_id = ?", trackingID).First(&shipping).Error; err != nil {
		return nil, err
	}
	return &shipping, nil
}

func (s *ShippingService) GetRates(origin, destination string, items []objects.Item) ([]objects.Rate, error) {
	if s.provider == nil {
		return nil, errors.New("shipping provider not set")
	}
	return s.provider.GetRates(origin, destination, items)
}

func (s *ShippingService) GetExpressMotorRates(origin, destination objects.LocationPrecise, items []objects.Item) ([]objects.Rate, error) {
	if s.provider == nil {
		return nil, errors.New("shipping provider not set")
	}
	return s.provider.GetExpressMotorRates(origin, destination, items)
}

func (s *ShippingService) TrackShipment(trackingID string) (*objects.TrackingStatus, error) {
	if s.provider == nil {
		return nil, errors.New("shipping provider not set")
	}
	shipping, err := s.GetShippingByTrackingID(trackingID)
	if err != nil {
		return nil, err
	}

	json.Unmarshal([]byte(shipping.TrackingData), &shipping.TrackingStatuses)

	status, err := s.provider.TrackShipment(trackingID)
	if err != nil {
		return nil, err
	}
	shipping.TrackingStatuses = status.History

	shipping.Status = status.Status

	b, err := json.Marshal(shipping.TrackingStatuses)
	if err != nil {
		return nil, err
	}
	shipping.TrackingData = string(b)
	if err := s.db.Save(&shipping).Error; err != nil {
		return nil, err
	}

	return &status, nil
}
