package provider

import "github.com/AMETORY/ametory-erp-modules/shared/objects"

type ShippingProvider interface {
	GetRates(origin, destination string, weight float64) ([]objects.Rate, error)
	GetExpressMotorRates(origin, destination objects.LocationPrecise) ([]objects.Rate, error)
	CreateShipment(orderID *string, destination string) (objects.Shipment, error)
	TrackShipment(trackingID string) (objects.TrackingStatus, error)
}
