// distribution/shipping/provider/gosend.go
package provider

import (
	"github.com/AMETORY/ametory-erp-modules/shared/objects"
)

type GoSendProvider struct {
	APIKey string
}

func NewGoSendProvider(apiKey string) *GoSendProvider {
	return &GoSendProvider{APIKey: apiKey}
}

func (g *GoSendProvider) GetRates(origin, destination string, items []objects.Item) ([]objects.Rate, error) {
	// Implementasi API GoSend untuk mendapatkan tarif
	return []objects.Rate{
		{Service: "SAME_DAY", Cost: 15000, ETA: "2 hours"},
		{Service: "INSTANT", Cost: 25000, ETA: "1 hour"},
	}, nil
}
func (g *GoSendProvider) GetExpressMotorRates(origin, destination objects.LocationPrecise, items []objects.Item) ([]objects.Rate, error) {
	// Implementasi API GoSend untuk mendapatkan tarif
	return []objects.Rate{
		{Service: "SAME_DAY", Cost: 15000, ETA: "2 hours", Distance: 4.3},
		{Service: "INSTANT", Cost: 25000, ETA: "1 hour", Distance: 4.3},
	}, nil
}

func (g *GoSendProvider) CreateDraftShipment(data interface{}) (objects.Shipment, error) {
	return objects.Shipment{
		Status: "Pending",
	}, nil
}
func (g *GoSendProvider) CreateShipment(data interface{}) (objects.Shipment, error) {
	// Implementasi API GoSend untuk membuat pengiriman
	return objects.Shipment{
		Status: "Pending",
	}, nil
}

func (g *GoSendProvider) TrackShipment(trackingID string) (objects.TrackingStatus, error) {
	// Implementasi API GoSend untuk melacak pengiriman
	return objects.TrackingStatus{
		TrackingID: &trackingID,
		Status:     "In Transit",
		Location:   "Jakarta",
	}, nil
}
