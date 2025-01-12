// distribution/shipping/provider/gosend.go
package provider

type GoSendProvider struct {
	APIKey string
}

func NewGoSendProvider(apiKey string) *GoSendProvider {
	return &GoSendProvider{APIKey: apiKey}
}

func (g *GoSendProvider) GetRates(origin, destination string, weight float64) ([]Rate, error) {
	// Implementasi API GoSend untuk mendapatkan tarif
	return []Rate{
		{Service: "SAME_DAY", Cost: 15000, ETA: "2 hours"},
		{Service: "INSTANT", Cost: 25000, ETA: "1 hour"},
	}, nil
}

func (g *GoSendProvider) CreateShipment(orderID uint, destination string) (Shipment, error) {
	// Implementasi API GoSend untuk membuat pengiriman
	return Shipment{
		TrackingID: "GS123456789",
		Status:     "Pending",
	}, nil
}

func (g *GoSendProvider) TrackShipment(trackingID string) (TrackingStatus, error) {
	// Implementasi API GoSend untuk melacak pengiriman
	return TrackingStatus{
		TrackingID: trackingID,
		Status:     "In Transit",
		Location:   "Jakarta",
	}, nil
}
