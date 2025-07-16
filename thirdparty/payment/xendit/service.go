package xendit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type XenditService struct {
	apiKey     string
	BaseURL    string
	apiVersion string
}

// NewXenditService creates a new instance of XenditService with the default values:
//
// - BaseURL: https://api.xendit.co
// - apiVersion: 2022-07-31
//
// You can then use the SetAPIKey or SetApiVersion method to set the API key or API version
// respectively.
func NewXenditService() *XenditService {
	return &XenditService{
		BaseURL:    "https://api.xendit.co",
		apiVersion: "2022-07-31",
	}
}

// SetAPIKey sets the API key for the Xendit service.
func (s *XenditService) SetAPIKey(apiKey string) {
	s.apiKey = apiKey
}

// SetApiVersion sets the API version for the Xendit service.
func (s *XenditService) SetApiVersion(version string) {
	s.apiVersion = version
}

// CreateQR creates a new QR code on the Xendit service.
//
// It takes an XenditQRrequest object as input and returns a pointer to an XenditQRResponse object.
// The object contains the created QR code's data, including the QR code string, QR code URL, and QR code ID.
// If there is an error, the method returns an error.
func (s *XenditService) CreateQR(req XenditQRrequest) (*XenditQRResponse, error) {

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequest("POST", s.BaseURL+"/qr_codes", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("api-version", s.apiVersion)
	httpReq.SetBasicAuth(s.apiKey, "")
	client := &http.Client{}
	client.Timeout = 30 * time.Second
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			err = cerr
		}
	}()

	var response XenditQRResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// GetQRByID retrieves a QR code from the Xendit service by its ID.
//
// It takes the ID of the QR code as a string and returns a pointer to an XenditQRResponse
// object containing the QR code's data. If there is an error, the method returns an error.
func (s *XenditService) GetQRByID(id string) (*XenditQRResponse, error) {
	httpReq, err := http.NewRequest("GET", fmt.Sprintf("%s/qr_codes/%s", s.BaseURL, id), nil)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("api-version", s.apiVersion)
	httpReq.SetBasicAuth(s.apiKey, "")
	client := &http.Client{}
	client.Timeout = 30 * time.Second
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			err = cerr
		}
	}()

	var response XenditQRResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// GetQRPayments retrieves a list of payments for a given QR code ID from the Xendit service.
//
// It takes a string parameter `qrID` representing the ID of the QR code and returns a slice of XenditQRPayment
// objects containing the payment data. If there is an error, the method returns an error.
func (s *XenditService) GetQRPayments(qrID string) ([]XenditQRPayment, error) {
	fmt.Printf("%s/qr_codes/%s/payments\n", s.BaseURL, qrID)
	httpReq, err := http.NewRequest("GET", fmt.Sprintf("%s/qr_codes/%s/payments", s.BaseURL, qrID), nil)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("api-version", s.apiVersion)
	httpReq.SetBasicAuth(s.apiKey, "")
	client := &http.Client{}
	client.Timeout = 30 * time.Second
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			err = cerr
		}
	}()

	var response map[string]any
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	var data = response["data"].([]any)

	dataPayments := []XenditQRPayment{}
	for _, v := range data {
		// t := reflect.TypeOf(v)
		// fmt.Println(v)
		dataPayments = append(dataPayments, XenditQRPayment{
			ID:          v.(map[string]any)["id"].(string),
			BusinessID:  v.(map[string]any)["business_id"].(string),
			Created:     v.(map[string]any)["created"].(string),
			Amount:      v.(map[string]any)["amount"].(float64),
			ChannelCode: v.(map[string]any)["channel_code"].(string),
			Currency:    v.(map[string]any)["currency"].(string),
			ExpiresAt:   v.(map[string]any)["expires_at"].(string),
			QRID:        v.(map[string]any)["qr_id"].(string),
			QRString:    v.(map[string]any)["qr_string"].(string),
			ReferenceID: v.(map[string]any)["reference_id"].(string),
			Status:      v.(map[string]any)["status"].(string),
			Type:        v.(map[string]any)["type"].(string),
		})

	}

	return dataPayments, nil
}
