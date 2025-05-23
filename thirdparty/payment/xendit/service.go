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

func NewXenditService() *XenditService {
	return &XenditService{
		BaseURL:    "https://api.xendit.co",
		apiVersion: "2022-07-31",
	}
}

func (s *XenditService) SetAPIKey(apiKey string) {
	s.apiKey = apiKey
}

func (s *XenditService) SetApiVersion(version string) {
	s.apiVersion = version
}

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
