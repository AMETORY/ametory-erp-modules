package provider

import "github.com/AMETORY/ametory-erp-modules/shared/objects"

type ShippingProvider interface {
	GetRates(origin, destination string, items []objects.Item) ([]objects.Rate, error)
	GetExpressMotorRates(origin, destination objects.LocationPrecise, items []objects.Item) ([]objects.Rate, error)
	CreateShipment(data interface{}) (objects.Shipment, error)
	CreateDraftShipment(data interface{}) (objects.Shipment, error)
	TrackShipment(trackingID string) (objects.TrackingStatus, error)
}
