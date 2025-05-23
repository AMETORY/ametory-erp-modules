package xendit

type XenditQRrequest struct {
	ReferenceID string  `json:"reference_id"`
	Type        string  `json:"type"`
	Currency    string  `json:"currency"`
	Amount      float64 `json:"amount"`
	ExpiresAt   string  `json:"expires_at"`
}

type XenditQRResponse struct {
	ReferenceID string  `json:"reference_id"`
	Type        string  `json:"type"`
	Currency    string  `json:"currency"`
	ChannelCode string  `json:"channel_code"`
	Amount      float64 `json:"amount"`
	ExpiresAt   string  `json:"expires_at"`
	Metadata    any     `json:"metadata"`
	BusinessID  string  `json:"business_id"`
	ID          string  `json:"id"`
	Created     string  `json:"created"`
	Updated     string  `json:"updated"`
	QRString    string  `json:"qr_string"`
	Status      string  `json:"status"`
}

type XenditQRPayment struct {
	Amount        float64 `json:"amount"`
	BusinessID    string  `json:"business_id"`
	ChannelCode   string  `json:"channel_code"`
	Created       string  `json:"created"`
	Currency      string  `json:"currency"`
	ExpiresAt     string  `json:"expires_at"`
	ID            string  `json:"id"`
	Metadata      any     `json:"metadata"`
	PaymentDetail struct {
		AccountDetails any `json:"account_details"`
		CustomerPan    any `json:"customer_pan"`
		MerchantPan    any `json:"merchant_pan"`
		Name           any `json:"name"`
		ReceiptID      any `json:"receipt_id"`
		Source         any `json:"source"`
	} `json:"payment_detail"`
	QRID        string `json:"qr_id"`
	QRString    string `json:"qr_string"`
	ReferenceID string `json:"reference_id"`
	Status      string `json:"status"`
	Type        string `json:"type"`
}

type XenditQRCode struct {
	ID         string `json:"id"`
	ExternalID string `json:"external_id"`
	QRString   string `json:"qr_string"`
	Type       string `json:"type"`
	Metadata   any    `json:"metadata"`
}

type XenditPaymentDetails struct {
	ReceiptID      string `json:"receipt_id"`
	Source         string `json:"source"`
	Name           string `json:"name"`
	AccountDetails any    `json:"account_details"`
}
type XenditGetQRPaymentResponse struct {
	Data    []XenditQRPayment `json:"data"`
	HasMore bool              `json:"has_more"`
}
