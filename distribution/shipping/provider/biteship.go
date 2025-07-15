package provider

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/shared/location"
	"github.com/AMETORY/ametory-erp-modules/shared/objects"
	"github.com/AMETORY/ametory-erp-modules/utils"
)

const (
	BiteShipAPIURL = "https://api.biteship.com"
)

type BiteShipProvider struct {
	APIKey string
}

// NewBiteShipProvider creates a new instance of BiteShipProvider.
// It requires an API key for BiteShip API.
func NewBiteShipProvider(apiKey string) *BiteShipProvider {
	return &BiteShipProvider{APIKey: apiKey}
}

// GetRates retrieves shipping rates for sending items from the origin to the destination.
// It returns a list of available shipping services with their respective costs and estimated time of arrival (ETA).
// Parameters:
//  - origin: the starting location for the shipment.
//  - destination: the endpoint location for the shipment.
//  - items: a list of items to be shipped, affecting the shipping cost and options.
// Returns:
//  - a slice of Rate objects, each representing a shipping option.
//  - an error if there is an issue retrieving the rates.

func (b *BiteShipProvider) GetRates(origin, destination string, items []objects.Item) ([]objects.Rate, error) {
	// Implementasi API GoSend untuk mendapatkan tarif
	return []objects.Rate{
		{Service: "SAME_DAY", Cost: 15000, ETA: "2 hours"},
		{Service: "INSTANT", Cost: 25000, ETA: "1 hour"},
	}, nil

}

// GetExpressMotorRates retrieves shipping rates for sending items from the origin to the destination using express motorbike services.
// It returns a list of available shipping services with their respective costs and estimated time of arrival (ETA).
// Parameters:
//   - origin: the starting location for the shipment.
//   - destination: the endpoint location for the shipment.
//   - items: a list of items to be shipped, affecting the shipping cost and options.
//
// Returns:
//   - a slice of Rate objects, each representing a shipping option.
//   - an error if there is an issue retrieving the rates.
func (b *BiteShipProvider) GetExpressMotorRates(origin, destination objects.LocationPrecise, items []objects.Item) ([]objects.Rate, error) {
	endpoint := "/v1/rates/couriers"
	// Implementasi API GoSend untuk mendapatkan tarif
	requestData := objects.ShipmentRateRequest{
		OriginLatitude:       origin.Latitude,
		OriginLongitude:      origin.Longitude,
		DestinationLatitude:  destination.Latitude,
		DestinationLongitude: destination.Longitude,
		Couriers:             "gojek,grab",
		Items:                items,
	}
	// utils.LogJson(requestData)
	body, err := json.Marshal(requestData)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", BiteShipAPIURL+endpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", b.APIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response objects.BiteShipRateResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	utils.LogJson(response)

	rates := make([]objects.Rate, len(response.Pricing))

	distance := location.Haversine(origin.Latitude, origin.Longitude, destination.Latitude, destination.Longitude)
	for i, rate := range response.Pricing {
		rates[i] = objects.Rate{
			Service:     rate.Type,
			Cost:        rate.Price,
			ETA:         rate.ShipmentDurationRange,
			CourierCode: rate.CourierCode,
			Distance:    distance,
		}
	}
	return rates, nil
}

// CreateDraftShipment creates a new shipment draft in the BiteShip API.
//
// It takes a BiteShipDraftShipmentRequest as input and attempts to save it to the BiteShip API.
// The function returns a pointer to a Shipment and an error. If the shipment is found, it is returned along with a nil
// error. If not found, or in case of a query error, the function returns a non-nil error.
func (b *BiteShipProvider) CreateDraftShipment(data interface{}) (objects.Shipment, error) {
	endpoint := "/v1/draft_orders"
	body, ok := data.(objects.BiteShipDraftShipmentRequest)
	if !ok {
		return objects.Shipment{}, errors.New("invalid data type")
	}

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return objects.Shipment{}, err
	}

	utils.LogJson(body)

	req, err := http.NewRequest("POST", BiteShipAPIURL+endpoint, bytes.NewBuffer(bodyJSON))
	if err != nil {
		return objects.Shipment{}, err
	}

	req.Header.Set("Authorization", b.APIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return objects.Shipment{}, err
	}
	defer resp.Body.Close()
	// Implementasi API GoSend untuk membuat pengiriman

	var response objects.BiteShipDraftOrderResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return objects.Shipment{}, err
	}

	utils.LogJson(response)

	var shipment objects.Shipment = objects.Shipment{
		Name:        response.Courier.Name,
		Phone:       response.Courier.Phone,
		Company:     response.Courier.Company,
		Type:        response.Courier.Type,
		Link:        response.Courier.Link,
		TrackingID:  response.Courier.TrackingID,
		WaybillID:   response.Courier.WaybillID,
		Insurance:   response.Courier.Insurance,
		RoutingCode: response.Courier.RoutingCode,
		Status:      response.Status,
	}

	return shipment, nil
}

// CreateShipment creates a new shipment order in the BiteShip API.
//
// It takes a BiteShipShipmentRequest as input and attempts to save it to the BiteShip API.
// The function returns a pointer to a Shipment and an error. If the shipment is found, it is returned along with a nil
// error. If not found, or in case of a query error, the function returns a non-nil error.
func (b *BiteShipProvider) CreateShipment(data interface{}) (objects.Shipment, error) {
	endpoint := "/v1/orders"
	body, ok := data.(objects.BiteShipShipmentRequest)
	if !ok {
		return objects.Shipment{}, errors.New("invalid data type")
	}

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return objects.Shipment{}, err
	}

	utils.LogJson(body)

	req, err := http.NewRequest("POST", BiteShipAPIURL+endpoint, bytes.NewBuffer(bodyJSON))
	if err != nil {
		return objects.Shipment{}, err
	}

	req.Header.Set("Authorization", b.APIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return objects.Shipment{}, err
	}
	defer resp.Body.Close()

	var response objects.BiteShipOrderResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return objects.Shipment{}, err
	}

	// utils.LogJson(response)
	if !response.Success {
		return objects.Shipment{}, errors.New(response.Error)
	}

	// Implementasi API GoSend untuk membuat pengiriman
	var shipment objects.Shipment = objects.Shipment{
		Name:        response.Courier.Name,
		Phone:       response.Courier.Phone,
		Company:     response.Courier.Company,
		Type:        response.Courier.Type,
		Link:        response.Courier.Link,
		TrackingID:  response.Courier.TrackingID,
		WaybillID:   response.Courier.WaybillID,
		Insurance:   response.Courier.Insurance,
		RoutingCode: response.Courier.RoutingCode,
		Status:      response.Status,
		ID:          response.ID,
		Price:       response.Price,
	}

	return shipment, nil
}

// TrackShipment tracks a shipment by its tracking ID and returns the latest status and information.
//
// It takes a tracking ID as input and returns a pointer to a TrackingStatus and an error. If the
// tracking ID is invalid or the request fails, the function returns an error. Otherwise, it returns
// the latest tracking information, including the tracking ID, waybill ID, status, link, history,
// and shipment information.
func (b *BiteShipProvider) TrackShipment(trackingID string) (objects.TrackingStatus, error) {
	endpoint := "/v1/trackings/" + trackingID
	req, err := http.NewRequest("GET", BiteShipAPIURL+endpoint, nil)
	if err != nil {
		return objects.TrackingStatus{}, err
	}

	req.Header.Set("Authorization", b.APIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return objects.TrackingStatus{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return objects.TrackingStatus{}, errors.New("failed to track shipment")
	}

	var response objects.BiteShipTrackingResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return objects.TrackingStatus{}, err
	}

	if !response.Success {
		return objects.TrackingStatus{}, errors.New(response.Error)
	}

	// utils.LogJson(response)
	// Implementasi API GoSend untuk melacak pengiriman
	return objects.TrackingStatus{
		TrackingID: &trackingID,
		WaybillID:  response.Courier.WaybillID,
		Status:     response.Status,
		Link:       response.Link,
		History:    response.History,
		Shipment:   response.Courier,
	}, nil
}
