package payment

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type EnvironmentType string

const (
	PROD    EnvironmentType = "PROD"
	STAGING EnvironmentType = "STAGING"
)
const (
	DefaultListEnableBank    = "002, 008, 009, 013, 022"
	DefaultListEnableEWallet = "shopeepay_ewallet"
)

type OyPayment struct {
	Username    string
	APIKey      string
	Environment EnvironmentType
	BaseURL     string
	// Add fields specific to OyPayment
}

func NewOyPayment(username, apiKey string, env EnvironmentType) *OyPayment {
	baseURL := "https://api-stg.oyindonesia.com"
	if env == PROD {
		baseURL = "https://api.oyindonesia.com"
	}
	return &OyPayment{
		Username: username,
		APIKey:   apiKey,
		BaseURL:  baseURL,
	}
}

type OyCreatePaymentLinkRequest struct {
	Description         string `json:"description"`
	PartnerTxID         string `json:"partner_tx_id"`
	Notes               string `json:"notes"`
	SenderName          string `json:"sender_name"`
	Amount              int    `json:"amount"`
	Email               string `json:"email"`
	PhoneNumber         string `json:"phone_number"`
	IsOpen              bool   `json:"is_open"` // If is_open = TRUE and the amount parameter is defined, then a payer can pay any amount (greater than IDR 10,000) up to the defined amount. And in the case that is_open=false, then the amount and partner_tx_id parameters must be defined.
	Step                string `json:"step"`
	IncludeAdminFee     bool   `json:"include_admin_fee"`             // Admin fee will be added to the specified amount or amount inputted by user if this parameter is set as TRUE.
	ListDisabledPayment string `json:"list_disabled_payment_methods"` // To configure payment methods to be disabled (e.g. VA, CREDIT_CARD, QRIS, EWALLET, BANK_TRANSFER)
	ListEnabledBanks    string `json:"list_enabled_banks"`            // List of eligible bank codes: "002" (BRI), "008" (Mandiri), "009" (BNI), "013" (Permata), "022" (CIMB), "213" (SMBC), "213" (BSI), and "014" (BCA).
	ListEnabledEwallet  string `json:"list_enabled_ewallet"`          // List of eligible e-wallet: "shopeepay_ewallet", "dana_ewallet", "linkaja_ewallet", "ovo_ewallet"
	Expiration          string `json:"expiration"`                    // To set the expiration of the payment link (yyyy-MM-dd HH:mm:ss) e.g. 2022-12-31 23:59:59
}

type OyCreatePaymentLinkResponse struct {
	PaymentLinkID string `json:"payment_link_id"`
	Message       string `json:"message"`
	EmailStatus   string `json:"email_status"`
	URL           string `json:"url"`
	Status        bool   `json:"status"`
}

func (o *OyPayment) CreatePaymentLink(data OyCreatePaymentLinkRequest) (*OyCreatePaymentLinkResponse, error) {
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
	var response *OyCreatePaymentLinkResponse
	err = json.NewDecoder(resp.Body).Decode(response)
	if err != nil {
		return nil, err
	}
	return response, nil
}
