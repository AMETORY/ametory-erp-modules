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

func NewBiteShipProvider(apiKey string) *BiteShipProvider {
	return &BiteShipProvider{APIKey: apiKey}
}

func (b *BiteShipProvider) GetRates(origin, destination string, items []objects.Item) ([]objects.Rate, error) {
	// Implementasi API GoSend untuk mendapatkan tarif
	return []objects.Rate{
		{Service: "SAME_DAY", Cost: 15000, ETA: "2 hours"},
		{Service: "INSTANT", Cost: 25000, ETA: "1 hour"},
	}, nil

}
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
	utils.LogJson(requestData)
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

	utils.LogJson(response)
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
