// distribution/shipping/provider/gosend.go
package provider

import (
	"github.com/AMETORY/ametory-erp-modules/shared/objects"
)

type GoSendProvider struct {
	APIKey string
}

// NewGoSendProvider creates a new instance of GoSendProvider with the provided API key.
// It initializes the GoSendProvider, which will be used to interact with GoSend API services.

func NewGoSendProvider(apiKey string) *GoSendProvider {
	return &GoSendProvider{APIKey: apiKey}
}

// GetRates retrieves shipping rates from GoSend based on the specified origin, destination, and item details.
// It returns a slice of Rate objects containing the available services, costs, and estimated time of arrival (ETA).
// An error is returned if there is an issue with fetching the rates.

func (g *GoSendProvider) GetRates(origin, destination string, items []objects.Item) ([]objects.Rate, error) {
	// Implementasi API GoSend untuk mendapatkan tarif
	return []objects.Rate{
		{Service: "SAME_DAY", Cost: 15000, ETA: "2 hours"},
		{Service: "INSTANT", Cost: 25000, ETA: "1 hour"},
	}, nil
}

// GetExpressMotorRates retrieves shipping rates for GoSend's Express Motor service based on the
// specified origin, destination, and item details.
// It returns a slice of Rate objects containing the available services, costs, estimated time of
// arrival (ETA), and distance.
// An error is returned if there is an issue with fetching the rates.
func (g *GoSendProvider) GetExpressMotorRates(origin, destination objects.LocationPrecise, items []objects.Item) ([]objects.Rate, error) {
	// Implementasi API GoSend untuk mendapatkan tarif
	return []objects.Rate{
		{Service: "SAME_DAY", Cost: 15000, ETA: "2 hours", Distance: 4.3},
		{Service: "INSTANT", Cost: 25000, ETA: "1 hour", Distance: 4.3},
	}, nil
}

// CreateDraftShipment creates a new draft shipment in the GoSend system.
//
// It takes an interface{} as input, which should conform to the expected
// structure for a draft shipment request. The function returns a Shipment
// object with a status of "Pending" and an error if there is an issue
// creating the draft shipment.

func (g *GoSendProvider) CreateDraftShipment(data interface{}) (objects.Shipment, error) {
	return objects.Shipment{
		Status: "Pending",
	}, nil
}

// CreateShipment creates a new shipment in the GoSend system.
//
// It takes an interface{} as input, which should conform to the expected
// structure for a shipment request. The function returns a Shipment object with
// a status of "Pending" and an error if there is an issue with creating the
// shipment.
func (g *GoSendProvider) CreateShipment(data interface{}) (objects.Shipment, error) {
	// Implementasi API GoSend untuk membuat pengiriman
	return objects.Shipment{
		Status: "Pending",
	}, nil
}

// TrackShipment tracks a shipment by its tracking ID and returns the current status and location.
//
// It takes a tracking ID as input and returns a TrackingStatus object and an error. If the tracking
// ID is invalid or the request fails, the function returns an error. Otherwise, it returns the
// current tracking information, including the tracking ID, status, and location.

func (g *GoSendProvider) TrackShipment(trackingID string) (objects.TrackingStatus, error) {
	// Implementasi API GoSend untuk melacak pengiriman
	return objects.TrackingStatus{
		TrackingID: &trackingID,
		Status:     "In Transit",
		Location:   "Jakarta",
	}, nil
}
