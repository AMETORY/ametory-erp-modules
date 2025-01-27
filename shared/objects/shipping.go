package objects

import "time"

type Rate struct {
	CourierCode string  `json:"courier_code"`
	Service     string  `json:"service"` // Jenis layanan (Reguler, Express, Same-Day, dll.)
	Cost        float64 `json:"cost"`    // Biaya pengiriman
	ETA         string  `json:"eta"`     // Perkiraan waktu pengiriman
	Distance    float64 `json:"distance"`
}

type Shipment struct {
	ID                string    `json:"id"`
	TrackingID        *string   `json:"tracking_id,omitempty"`
	WaybillID         *string   `json:"waybill_id,omitempty"`
	Company           string    `json:"company"`         // "sicepat"
	Name              *string   `json:"name,omitempty"`  // Deprecated
	Phone             *string   `json:"phone,omitempty"` // Deprecated
	DriverName        *string   `json:"driver_name,omitempty"`
	DriverPhone       *string   `json:"driver_phone,omitempty"`
	DriverPhotoURL    *string   `json:"driver_photo_url,omitempty"`
	DriverPlateNumber *string   `json:"driver_plate_number,omitempty"`
	Type              string    `json:"type"` // "reg"
	Link              *string   `json:"link,omitempty"`
	Insurance         Insurance `json:"insurance"`
	RoutingCode       *string   `json:"routing_code,omitempty"`
	Status            string    `json:"status"`
	Price             float64   `json:"price"`
}

type Insurance struct {
	Amount int    `json:"amount"`
	Fee    int    `json:"fee"`
	Note   string `json:"note"`
}

type TrackingStatus struct {
	Date       time.Time `json:"date"`        // Tanggal status diperbarui
	TrackingID *string   `json:"tracking_id"` // Nomor resi
	Status     string    `json:"status"`      // Status terbaru
	Location   string    `json:"location"`    // Lokasi terbaru
	WaybillID  *string   `json:"waybill_id"`
	Link       string    `json:"link"`
	History    []History `json:"history"`
}

type LocationPrecise struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Notes     string  `json:"notes"`
}

type Item struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Value       float64 `json:"value"`
	Length      int     `json:"length"`
	Width       int     `json:"width"`
	Height      int     `json:"height"`
	Weight      int     `json:"weight"`
	Quantity    int     `json:"quantity"`
}
