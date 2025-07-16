package oy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/shared/objects"
	"github.com/AMETORY/ametory-erp-modules/utils"
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

// NewOyPaymentService creates a new instance of OyPaymentService with the given username,
// api key, and environment. It sets the base URL for the Oy API based on the environment.
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

// CreatePaymentVA creates a virtual account payment request using the Oy API.
// It takes an interface{} as the dataPayment parameter, which should be of type
// OyCreatePaymentVARequest. The function marshals the request data into JSON format
// and sends a POST request to the Oy API's "generate-static-va" endpoint. It sets
// the necessary headers for authentication. If the request is successful, it decodes
// the JSON response into an OyCreatePaymentVAResponse struct and returns it. In case
// of an error during marshaling, request creation, API call, or response decoding,
// an error is returned.
func (o *OyPaymentService) CreatePaymentVA(dataPayment interface{}) (interface{}, error) {
	data, ok := dataPayment.(OyCreatePaymentVARequest)
	if !ok {
		return nil, fmt.Errorf("invalid data type")
	}
	utils.LogJson(data)
	client := &http.Client{}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", o.BaseURL+"/api/generate-static-va", bytes.NewBuffer(jsonData))
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
	var response OyCreatePaymentVAResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// CreatePaymentEWallet creates an e-wallet payment request using the Oy API.
// It takes an interface{} as the dataPayment parameter, which should be of type
// OyCreatePaymentEWalletRequest. The function marshals the request data into JSON format
// and sends a POST request to the Oy API's "e-wallet-aggregator/create-transaction" endpoint. It sets
// the necessary headers for authentication. If the request is successful, it decodes
// the JSON response into an OyCreatePaymentEWalletResponse struct and returns it. In case
// of an error during marshaling, request creation, API call, or response decoding,
// an error is returned.
func (o *OyPaymentService) CreatePaymentEWallet(dataPayment interface{}) (interface{}, error) {
	data, ok := dataPayment.(OyCreatePaymentEWalletRequest)
	if !ok {
		return nil, fmt.Errorf("invalid data type")
	}
	utils.LogJson(data)
	client := &http.Client{}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", o.BaseURL+"/api/e-wallet-aggregator/create-transaction", bytes.NewBuffer(jsonData))
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
	var response OyCreatePaymentEWalletResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// CreatePaymentLink creates a payment link request using the Oy API.
// It takes an interface{} as the dataPayment parameter, which should be of type
// OyCreatePaymentLinkRequest. The function marshals the request data into JSON format
// and sends a POST request to the Oy API's "payment-checkout/create-v2" endpoint. It sets
// the necessary headers for authentication. If the request is successful, it decodes
// the JSON response into an OyCreatePaymentLinkResponse struct and returns it. In case
// of an error during marshaling, request creation, API call, or response decoding,
// an error is returned.
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
	utils.LogJson(data)
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

// DetailPaymentVA retrieves details of a virtual account payment using the Oy API.
// It takes a variadic interface{} parameter, where the first element should be a slice of interfaces
// containing the virtual account ID as a string. The function sends a GET request to the Oy API's
// "static-virtual-account" endpoint to fetch the payment details. It sets the necessary headers
// for authentication. If the request is successful, it decodes the JSON response into an
// OyCreatePaymentVAResponse struct and returns it. In case of an error during request
// creation, API call, or response decoding, an error is returned.
func (o *OyPaymentService) DetailPaymentVA(data ...interface{}) (interface{}, error) {
	if len(data) < 1 {
		return nil, fmt.Errorf("invalid data")
	}
	var ids = data[0].([]interface{})
	id := ids[0].(string)
	client := &http.Client{}
	fmt.Println("GET  VA STATUS", fmt.Sprintf("%s/api/static-virtual-account/%s", o.BaseURL, id))

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/static-virtual-account/%s", o.BaseURL, id), nil)
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
	var response OyCreatePaymentVAResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	// utils.LogJson(response)
	return response, nil

}

// DetailPaymentEWallet retrieves details of an e-wallet payment using the Oy API.
// It takes a variadic interface{} parameter, where the first element should be a slice of interfaces
// containing the transaction ID as a string. The function sends a POST request to the Oy API's
// "e-wallet-aggregator/check-status" endpoint to fetch the payment details. It sets the necessary headers
// for authentication. If the request is successful, it decodes the JSON response into an
// OyCreatePaymentEWalletResponse struct and returns it. In case of an error during request
// creation, API call, or response decoding, an error is returned.
func (o *OyPaymentService) DetailPaymentEWallet(data ...interface{}) (interface{}, error) {
	if len(data) < 1 {
		return nil, fmt.Errorf("invalid data")
	}
	var ids = data[0].([]interface{})
	id := ids[0].(string)
	jsonData, err := json.Marshal(map[string]interface{}{
		"partner_trx_id": id,
	})
	if err != nil {
		return nil, err
	}
	client := &http.Client{}
	// fmt.Println("GET  VA STATUS", fmt.Sprintf("%s/api/static-virtual-account/%s", o.BaseURL, id))

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/e-wallet-aggregator/check-status", o.BaseURL), bytes.NewBuffer(jsonData))
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
	var response OyCreatePaymentEWalletResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	// utils.LogJson(response)
	return response, nil

}

// DetailPayment retrieves a payment status from the Oy API.
// The function takes a variadic interface{} parameter, where the first element should be a slice of interfaces
// containing the partner transaction ID as a string and a boolean indicating whether to send a callback or not.
// The function sends a GET request to the Oy API's "payment-checkout/status" endpoint to fetch the payment status. It sets
// the necessary headers for authentication. If the request is successful, it decodes the JSON response into an
// OyPaymentResponse struct and returns it. In case of an error during request creation, API call, or response decoding,
// an error is returned.
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
