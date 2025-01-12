package provider

import "time"

type ShippingProvider interface {
	GetRates(origin, destination string, weight float64) ([]Rate, error)
	CreateShipment(orderID uint, destination string) (Shipment, error)
	TrackShipment(trackingID string) (TrackingStatus, error)
}

type Rate struct {
	Service string  `json:"service"` // Jenis layanan (Reguler, Express, Same-Day, dll.)
	Cost    float64 `json:"cost"`    // Biaya pengiriman
	ETA     string  `json:"eta"`     // Perkiraan waktu pengiriman
}

type Shipment struct {
	TrackingID string `json:"tracking_id"` // Nomor resi
	Status     string `json:"status"`      // Status pengiriman
}

type TrackingStatus struct {
	Date       time.Time `json:"date"`        // Tanggal status diperbarui
	TrackingID string    `json:"tracking_id"` // Nomor resi
	Status     string    `json:"status"`      // Status terbaru
	Location   string    `json:"location"`    // Lokasi terbaru
}
