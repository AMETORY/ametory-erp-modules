package google

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
)

type GoogleAPIService struct {
	ctx    *context.ERPContext
	apiKey string
}

func NewGoogleAPIService(ctx *context.ERPContext, apiKey string) *GoogleAPIService {
	return &GoogleAPIService{ctx: ctx, apiKey: apiKey}
}

func (s *GoogleAPIService) SearchPlaceByCoordinate(latitude float64, longitude float64, maxResult int, radius float64) (*PlacesResponse, error) {
	url := "https://places.googleapis.com/v1/places:searchNearby"

	request := map[string]interface{}{
		"maxResultCount": maxResult,
		"locationRestriction": map[string]interface{}{
			"circle": map[string]interface{}{
				"center": map[string]float64{
					"latitude":  latitude,
					"longitude": longitude,
				},
				"radius": radius,
			},
		},
	}
	reqBody, err := json.Marshal(request)
	if err != nil {
		fmt.Println("Error marshalling request:", err)
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return nil, err
	}
	if err != nil {
		fmt.Println("Error creating request:", err)
		return nil, err
	}

	fmt.Println("API KEY:", s.apiKey)
	// Menambahkan headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Goog-Api-Key", s.apiKey)
	req.Header.Set("X-Goog-FieldMask", "places.displayName,places.formattedAddress,places.location")

	// Mengirim request menggunakan HTTP client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return nil, err
	}
	defer resp.Body.Close()

	// Membaca response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return nil, err
	}

	// Menampilkan response
	fmt.Println("Response Status:", resp.Status)
	fmt.Println("Response Body:", string(body))

	var response PlacesResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Println("Error unmarshalling response:", err)
		return nil, err
	}
	return &response, nil
}
func (s *GoogleAPIService) SearchPlace(keyword string) (*PlacesResponse, error) {
	url := "https://places.googleapis.com/v1/places:searchText"

	request := map[string]string{
		"textQuery": keyword,
	}

	// Membuat HTTP request
	reqBody, err := json.Marshal(request)
	if err != nil {
		fmt.Println("Error marshalling request:", err)
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return nil, err
	}
	if err != nil {
		fmt.Println("Error creating request:", err)
		return nil, err
	}

	fmt.Println("API KEY:", s.apiKey)
	// Menambahkan headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Goog-Api-Key", s.apiKey)
	req.Header.Set("X-Goog-FieldMask", "places.displayName,places.formattedAddress,places.location")

	// Mengirim request menggunakan HTTP client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return nil, err
	}
	defer resp.Body.Close()

	// Membaca response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return nil, err
	}

	// Menampilkan response
	fmt.Println("Response Status:", resp.Status)
	fmt.Println("Response Body:", string(body))

	var response PlacesResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Println("Error unmarshalling response:", err)
		return nil, err
	}
	return &response, nil
}

type PlacesResponse struct {
	Places []Place `json:"places"`
}

type Place struct {
	FormattedAddress string   `json:"formattedAddress"`
	Location         Location `json:"location"`
	DisplayName      struct {
		Text string `json:"text"`
	} `json:"displayName"`
}

type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}
