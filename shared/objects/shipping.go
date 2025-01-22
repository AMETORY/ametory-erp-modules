package objects

import "time"

type Rate struct {
	Service  string  `json:"service"` // Jenis layanan (Reguler, Express, Same-Day, dll.)
	Cost     float64 `json:"cost"`    // Biaya pengiriman
	ETA      string  `json:"eta"`     // Perkiraan waktu pengiriman
	Distance float64 `json:"distance"`
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

type LocationPrecise struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Notes     string  `json:"notes"`
}
