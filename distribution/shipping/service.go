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

// NewShippingService creates a new instance of ShippingService with the given database connection and context.
func NewShippingService(db *gorm.DB, ctx *context.ERPContext) *ShippingService {
	return &ShippingService{db: db, ctx: ctx}
}

// Migrate migrates the database schema of the Shipping module.
//
// It uses the `AutoMigrate` method of GORM to create the database tables
// for the Shipping module, if they do not already exist.
//
// The function returns an error if the migration fails.
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&models.ShippingModel{})
}

// SetProvider sets the shipping provider for this ShippingService.
//
// The provider parameter must be an implementation of the ShippingProvider
// interface.
//
// The function does not return an error. If the provider is not set, the
// ShippingService will not be able to perform any operations.
func (s *ShippingService) SetProvider(provider provider.ShippingProvider) {
	s.provider = provider
}

// CreateDraftShipment creates a new shipment draft in the provider's system.
//
// It takes an interface{} as input, which should conform to the expected
// structure for a draft shipment request. The function returns a Shipment
// object with a status of "Pending" and an error if there is an issue
// creating the draft shipment.
func (s *ShippingService) CreateDraftShipment(data interface{}) (objects.Shipment, error) {
	return s.provider.CreateDraftShipment(data)
}

// CreateShipment creates a new shipment order in the provider's system.
//
// It takes an interface{} as input, which should conform to the expected
// structure for a shipment request. The function returns a Shipment object
// with a status of "Pending" and an error if there is an issue creating the
// shipment.
func (s *ShippingService) CreateShipment(data interface{}) (objects.Shipment, error) {
	return s.provider.CreateShipment(data)
}

// GetShippingByOrderID retrieves a ShippingModel from the database by
// order ID.
//
// It returns a pointer to a ShippingModel and an error if the operation
// fails. If the shipping is found, it is returned along with a nil error.
// If not found, or in case of a query error, the function returns a non-nil
// error.
func (s *ShippingService) GetShippingByOrderID(orderID string) (*models.ShippingModel, error) {
	var shipping models.ShippingModel
	if err := s.db.Where("order_id = ?", orderID).First(&shipping).Error; err != nil {
		return nil, err
	}
	return &shipping, nil
}

// GetShippingByID retrieves a ShippingModel from the database by its ID.
//
// It returns a pointer to a ShippingModel and an error if the operation
// fails. If the shipping is found, it is returned along with a nil error.
// If not found, or in case of a query error, the function returns a non-nil
// error.
func (s *ShippingService) GetShippingByID(ID string) (*models.ShippingModel, error) {
	var shipping models.ShippingModel
	if err := s.db.Where("id = ?", ID).First(&shipping).Error; err != nil {
		return nil, err
	}
	return &shipping, nil
}

// GetShippingByTrackingID retrieves a ShippingModel from the database by its
// tracking ID.
//
// It returns a pointer to a ShippingModel and an error if the operation
// fails. If the shipping is found, it is returned along with a nil error.
// If not found, or in case of a query error, the function returns a non-nil
// error.
func (s *ShippingService) GetShippingByTrackingID(trackingID string) (*models.ShippingModel, error) {
	var shipping models.ShippingModel
	if err := s.db.Where("tracking_id = ?", trackingID).First(&shipping).Error; err != nil {
		return nil, err
	}
	return &shipping, nil
}

// GetRates retrieves shipping rates from the configured provider.
//
// It takes an origin and destination location as strings, and a slice of Item objects.
// It returns a slice of Rate objects containing the available services, costs, estimated time of
// arrival (ETA), and distances.
// An error is returned if there is an issue with fetching the rates, or if the shipping provider
// is not set.
func (s *ShippingService) GetRates(origin, destination string, items []objects.Item) ([]objects.Rate, error) {
	if s.provider == nil {
		return nil, errors.New("shipping provider not set")
	}
	return s.provider.GetRates(origin, destination, items)
}

// GetExpressMotorRates retrieves shipping rates for sending items from the origin to the destination using express motorbike services.
//
// It takes an origin and destination location as LocationPrecise objects, and a slice of Item objects.
// It returns a slice of Rate objects containing the available services, costs, and estimated time of arrival (ETA).
// An error is returned if there is an issue with fetching the rates, or if the shipping provider is not set.
func (s *ShippingService) GetExpressMotorRates(origin, destination objects.LocationPrecise, items []objects.Item) ([]objects.Rate, error) {
	if s.provider == nil {
		return nil, errors.New("shipping provider not set")
	}
	return s.provider.GetExpressMotorRates(origin, destination, items)
}

// TrackShipment tracks a shipment by its tracking ID and updates the shipping details in the database.
//
// It takes a tracking ID as input and returns a pointer to a TrackingStatus object and an error.
// If the shipping provider is not set, it returns an error. If the shipment is found, its tracking data is
// unmarshalled, and the latest tracking information is fetched from the provider. The function updates the
// shipment's tracking statuses and courier details in the database. If the tracking ID is invalid or the request
// fails, an error is returned.

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
	shipping.Courier = status.Shipment

	shipping.Status = status.Status

	b, err := json.Marshal(shipping.TrackingStatuses)
	if err != nil {
		return nil, err
	}
	shipping.TrackingData = string(b)

	c, err := json.Marshal(shipping.Courier)
	if err != nil {
		return nil, err
	}
	shipping.CourierData = string(c)

	if err := s.db.Save(&shipping).Error; err != nil {
		return nil, err
	}

	return &status, nil
}
