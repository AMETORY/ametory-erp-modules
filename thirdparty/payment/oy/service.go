package oy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/shared/objects"
)

const (
	DefaultListEnableBank    = "002, 008, 009, 013, 022"
	DefaultListEnableEWallet = "shopeepay_ewallet"
)

type OyPaymentService struct {
	Username    string
	APIKey      string
	Environment objects.EnvironmentType
	BaseURL     string
	// Add fields specific to OyPaymentService
}

func NewOyPaymentService(username, apiKey string, env objects.EnvironmentType) *OyPaymentService {
	baseURL := "https://api-stg.oyindonesia.com"
	if env == objects.PROD {
		baseURL = "https://api.oyindonesia.com"
	}
	return &OyPaymentService{
		Username: username,
		APIKey:   apiKey,
		BaseURL:  baseURL,
	}
}

func (o *OyPaymentService) CreatePaymentLink(dataPayment interface{}) (interface{}, error) {
	data, ok := dataPayment.(OyCreatePaymentLinkRequest)
	if !ok {
		return nil, fmt.Errorf("invalid data type")
	}
	if data.ListEnabledBanks == "" {
		data.ListEnabledBanks = DefaultListEnableBank
	}
	if data.ListEnabledEwallet == "" {
		data.ListEnabledEwallet = DefaultListEnableEWallet
	}
	client := &http.Client{}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", o.BaseURL+"/api/payment-checkout/create-v2", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Oy-Username", o.Username)
	req.Header.Set("X-Api-Key", o.APIKey)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	// fmt.Println(resp.Body)
	var response OyCreatePaymentLinkResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (o *OyPaymentService) DetailPayment(data ...interface{}) (interface{}, error) {

	if len(data) < 1 {
		return nil, fmt.Errorf("invalid data")
	}
	var newData = data[0].([]interface{})

	partnerTxID := newData[0].(string)
	sendCallBack := false
	sendCallBack = newData[1].(bool)

	fmt.Println(partnerTxID, sendCallBack)

	client := &http.Client{}
	sendCallBackStr := "false"
	if sendCallBack {
		sendCallBackStr = "true"
	}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/payment-checkout/status?partner_tx_id=%s&send_callback=%s", o.BaseURL, partnerTxID, sendCallBackStr), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Oy-Username", o.Username)
	req.Header.Set("X-Api-Key", o.APIKey)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var response OyPaymentResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}
	return response, nil
}
