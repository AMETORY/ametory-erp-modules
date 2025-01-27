package objects

import "time"

type Coordinate struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type BiteShipDraftShipmentRequest struct {
	OriginContactName       string `json:"origin_contact_name"`
	OriginContactPhone      string `json:"origin_contact_phone"`
	OriginAddress           string `json:"origin_address"`
	OriginNote              string `json:"origin_note"`
	OriginPostalCode        int    `json:"origin_postal_code"`
	DestinationContactName  string `json:"destination_contact_name"`
	DestinationContactPhone string `json:"destination_contact_phone"`
	DestinationContactEmail string `json:"destination_contact_email"`
	DestinationAddress      string `json:"destination_address"`
	DestinationPostalCode   int    `json:"destination_postal_code"`
	DestinationNote         string `json:"destination_note"`
	CourierCompany          string `json:"courier_company"`
	CourierType             string `json:"courier_type"`
	DeliveryType            string `json:"delivery_type"`
	OrderNote               string `json:"order_note"`
	Items                   []Item `json:"items"`
}

type BiteShipShipmentRequest struct {
	ShipperContactName      string                 `json:"shipper_contact_name"`
	ShipperContactPhone     string                 `json:"shipper_contact_phone"`
	ShipperContactEmail     string                 `json:"shipper_contact_email"`
	ShipperOrganization     string                 `json:"shipper_organization"`
	OriginContactName       string                 `json:"origin_contact_name"`
	OriginContactPhone      string                 `json:"origin_contact_phone"`
	OriginAddress           string                 `json:"origin_address"`
	OriginNote              string                 `json:"origin_note"`
	OriginCoordinate        Coordinate             `json:"origin_coordinate"`
	DestinationContactName  string                 `json:"destination_contact_name"`
	DestinationContactPhone string                 `json:"destination_contact_phone"`
	DestinationContactEmail string                 `json:"destination_contact_email"`
	DestinationAddress      string                 `json:"destination_address"`
	DestinationNote         string                 `json:"destination_note"`
	DestinationCoordinate   Coordinate             `json:"destination_coordinate"`
	CourierCompany          string                 `json:"courier_company"`
	CourierType             string                 `json:"courier_type"`
	CourierInsurance        int                    `json:"courier_insurance"`
	DeliveryType            string                 `json:"delivery_type"`
	OrderNote               string                 `json:"order_note"`
	Metadata                map[string]interface{} `json:"metadata"`
	Items                   []Item                 `json:"items"`
}

type BiteShipOrderResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Message string `json:"message"`
	Object  string `json:"object"`
	ID      string `json:"id"`
	Shipper struct {
		Name         string `json:"name"`
		Email        string `json:"email"`
		Phone        string `json:"phone"`
		Organization string `json:"organization"`
	} `json:"shipper"`
	Origin struct {
		ContactName  string     `json:"contact_name"`
		ContactPhone string     `json:"contact_phone"`
		Coordinate   Coordinate `json:"coordinate"`
		Address      string     `json:"address"`
		Note         string     `json:"note"`
		PostalCode   int        `json:"postal_code"`
	} `json:"origin"`
	Destination struct {
		ContactName     string `json:"contact_name"`
		ContactPhone    string `json:"contact_phone"`
		ContactEmail    string `json:"contact_email"`
		Address         string `json:"address"`
		Note            string `json:"note"`
		ProofOfDelivery struct {
			Use  bool   `json:"use"`
			Fee  int    `json:"fee"`
			Note string `json:"note"`
			Link string `json:"link"`
		} `json:"proof_of_delivery"`
		CashOnDelivery struct {
			ID     string `json:"id"`
			Amount int    `json:"amount"`
			Fee    int    `json:"fee"`
			Note   string `json:"note"`
			Type   string `json:"type"`
		} `json:"cash_on_delivery"`
		Coordinate Coordinate `json:"coordinate"`
		PostalCode int        `json:"postal_code"`
	} `json:"destination"`
	Courier  Shipment `json:"courier"`
	Delivery struct {
		Datetime     string  `json:"datetime"`
		Note         string  `json:"note"`
		Type         string  `json:"type"`
		Distance     float64 `json:"distance"`
		DistanceUnit string  `json:"distance_unit"`
	} `json:"delivery"`
	ReferenceID string `json:"reference_id"`
	Items       []struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		SKU         string `json:"sku"`
		Value       int    `json:"value"`
		Quantity    int    `json:"quantity"`
		Length      int    `json:"length"`
		Width       int    `json:"width"`
		Height      int    `json:"height"`
		Weight      int    `json:"weight"`
	} `json:"items"`
	Extra    []interface{}          `json:"extra"`
	Price    float64                `json:"price"`
	Metadata map[string]interface{} `json:"metadata"`
	Note     string                 `json:"note"`
	Status   string                 `json:"status"`
}

type ShipmentRateRequest struct {
	OriginLatitude       float64 `json:"origin_latitude"`
	OriginLongitude      float64 `json:"origin_longitude"`
	DestinationLatitude  float64 `json:"destination_latitude"`
	DestinationLongitude float64 `json:"destination_longitude"`
	Couriers             string  `json:"couriers"`
	Items                []Item  `json:"items"`
}

type BiteShipRateResponse struct {
	Success bool   `json:"success"`
	Object  string `json:"object"`
	Message string `json:"message"`
	Code    int    `json:"code"`
	Origin  struct {
		LocationID      string  `json:"location_id"`
		Latitude        float64 `json:"latitude"`
		Longitude       float64 `json:"longitude"`
		PostalCode      string  `json:"postal_code"`
		CountryName     string  `json:"country_name"`
		CountryCode     string  `json:"country_code"`
		Administrative1 struct {
			Name string `json:"name"`
			Type string `json:"type"`
		} `json:"administrative_division_level_1"`
		Administrative2 struct {
			Name string `json:"name"`
			Type string `json:"type"`
		} `json:"administrative_division_level_2"`
		Administrative3 struct {
			Name string `json:"name"`
			Type string `json:"type"`
		} `json:"administrative_division_level_3"`
		Administrative4 struct {
			Name string `json:"name"`
			Type string `json:"type"`
		} `json:"administrative_division_level_4"`
		Address string `json:"address"`
	} `json:"origin"`
	Destination struct {
		LocationID      string  `json:"location_id"`
		Latitude        float64 `json:"latitude"`
		Longitude       float64 `json:"longitude"`
		PostalCode      string  `json:"postal_code"`
		CountryName     string  `json:"country_name"`
		CountryCode     string  `json:"country_code"`
		Administrative1 struct {
			Name string `json:"name"`
			Type string `json:"type"`
		} `json:"administrative_division_level_1"`
		Administrative2 struct {
			Name string `json:"name"`
			Type string `json:"type"`
		} `json:"administrative_division_level_2"`
		Administrative3 struct {
			Name string `json:"name"`
			Type string `json:"type"`
		} `json:"administrative_division_level_3"`
		Administrative4 struct {
			Name string `json:"name"`
			Type string `json:"type"`
		} `json:"administrative_division_level_4"`
		Address string `json:"address"`
	} `json:"destination"`
	Pricing []struct {
		AvailableForCashOnDelivery   bool     `json:"available_for_cash_on_delivery"`
		AvailableForProofOfDelivery  bool     `json:"available_for_proof_of_delivery"`
		AvailableForInstantWaybillID bool     `json:"available_for_instant_waybill_id"`
		AvailableForInsurance        bool     `json:"available_for_insurance"`
		AvailableCollectionMethod    []string `json:"available_collection_method"`
		Company                      string   `json:"company"`
		CourierName                  string   `json:"courier_name"`
		CourierCode                  string   `json:"courier_code"`
		CourierServiceName           string   `json:"courier_service_name"`
		CourierServiceCode           string   `json:"courier_service_code"`
		Description                  string   `json:"description"`
		Duration                     string   `json:"duration"`
		ShipmentDurationRange        string   `json:"shipment_duration_range"`
		ShipmentDurationUnit         string   `json:"shipment_duration_unit"`
		ServiceType                  string   `json:"service_type"`
		ShippingType                 string   `json:"shipping_type"`
		Price                        float64  `json:"price"`
		Type                         string   `json:"type"`
	} `json:"pricing"`
}

type BiteShipDraftOrderResponse struct {
	Success bool   `json:"success"`
	Code    int    `json:"code"`
	Object  string `json:"object"`
	ID      string `json:"id"`
	Origin  struct {
		AreaID       string `json:"area_id"`
		Address      string `json:"address"`
		Note         string `json:"note"`
		ContactName  string `json:"contact_name"`
		ContactPhone string `json:"contact_phone"`
		ContactEmail string `json:"contact_email"`
		Coordinate   struct {
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
		} `json:"coordinate"`
		ProvinceName     string `json:"province_name"`
		CityName         string `json:"city_name"`
		DistrictName     string `json:"district_name"`
		PostalCode       int    `json:"postal_code"`
		CollectionMethod string `json:"collection_method"`
	} `json:"origin"`
	Destination struct {
		AreaID       string `json:"area_id"`
		Address      string `json:"address"`
		Note         string `json:"note"`
		ContactName  string `json:"contact_name"`
		ContactPhone string `json:"contact_phone"`
		ContactEmail string `json:"contact_email"`
		Coordinate   struct {
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
		} `json:"coordinate"`
		ProvinceName    string `json:"province_name"`
		CityName        string `json:"city_name"`
		DistrictName    string `json:"district_name"`
		PostalCode      int    `json:"postal_code"`
		ProofOfDelivery struct {
			Use  bool   `json:"use"`
			Fee  int    `json:"fee"`
			Note string `json:"note"`
			Link string `json:"link"`
		} `json:"proof_of_delivery"`
		CashOnDelivery struct {
			PaymentMethod string  `json:"payment_method"`
			Amount        float64 `json:"amount"`
			Note          string  `json:"note"`
			Type          string  `json:"type"`
		} `json:"cash_on_delivery"`
	} `json:"destination"`
	Courier  Shipment `json:"courier"`
	Delivery struct {
		Type         string    `json:"type"`
		Datetime     time.Time `json:"datetime"`
		Note         string    `json:"note"`
		Distance     float64   `json:"distance"`
		DistanceUnit string    `json:"distance_unit"`
	} `json:"delivery"`
	Extra    []string    `json:"extra"`
	Tags     []string    `json:"tags"`
	Metadata interface{} `json:"metadata"`
	Items    []struct {
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Value       float64 `json:"value"`
		Quantity    int     `json:"quantity"`
		Height      int     `json:"height"`
		Width       int     `json:"width"`
		Length      int     `json:"length"`
		Weight      int     `json:"weight"`
	} `json:"items"`
	Price       float64   `json:"price"`
	Status      string    `json:"status"`
	ReferenceID string    `json:"reference_id"`
	InvoiceID   string    `json:"invoice_id"`
	UserID      string    `json:"user_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	PlacedAt    time.Time `json:"placed_at"`
	ReadyAt     time.Time `json:"ready_at"`
	ConfirmedAt time.Time `json:"confirmed_at"`
	DeletedAt   time.Time `json:"deleted_at"`
}

type BiteShipTrackingResponse struct {
	Success   bool     `json:"success"`
	Error     string   `json:"error"`
	Message   string   `json:"message"`
	Object    string   `json:"object"`
	ID        string   `json:"id"`
	WaybillID string   `json:"waybill_id"`
	Courier   Shipment `json:"courier"`
	Origin    struct {
		ContactName string `json:"contact_name"`
		Address     string `json:"address"`
	} `json:"origin"`
	Destination struct {
		ContactName string `json:"contact_name"`
		Address     string `json:"address"`
	} `json:"destination"`
	History []History `json:"history"`
	Link    string    `json:"link"`
	OrderID string    `json:"order_id"`
	Status  string    `json:"status"`
}

type History struct {
	Note        string    `json:"note"`
	ServiceType string    `json:"service_type"`
	UpdatedAt   time.Time `json:"updated_at"`
	Status      string    `json:"status"`
}
