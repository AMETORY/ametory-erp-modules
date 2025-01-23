package oy

type OyCreatePaymentLinkRequest struct {
	Description         string  `json:"description"`
	PartnerTxID         string  `json:"partner_tx_id"`
	Notes               string  `json:"notes"`
	SenderName          string  `json:"sender_name"`
	Amount              float64 `json:"amount"`
	Email               string  `json:"email"`
	PhoneNumber         string  `json:"phone_number"`
	IsOpen              bool    `json:"is_open"` // If is_open = TRUE and the amount parameter is defined, then a payer can pay any amount (greater than IDR 10,000) up to the defined amount. And in the case that is_open=false, then the amount and partner_tx_id parameters must be defined.
	Step                string  `json:"step"`
	IncludeAdminFee     bool    `json:"include_admin_fee"`             // Admin fee will be added to the specified amount or amount inputted by user if this parameter is set as TRUE.
	ListDisabledPayment string  `json:"list_disabled_payment_methods"` // To configure payment methods to be disabled (e.g. VA, CREDIT_CARD, QRIS, EWALLET, BANK_TRANSFER)
	ListEnabledBanks    string  `json:"list_enabled_banks"`            // List of eligible bank codes: "002" (BRI), "008" (Mandiri), "009" (BNI), "013" (Permata), "022" (CIMB), "213" (SMBC), "213" (BSI), and "014" (BCA).
	ListEnabledEwallet  string  `json:"list_enabled_ewallet"`          // List of eligible e-wallet: "shopeepay_ewallet", "dana_ewallet", "linkaja_ewallet", "ovo_ewallet"
	Expiration          string  `json:"expiration"`                    // To set the expiration of the payment link (yyyy-MM-dd HH:mm:ss) e.g. 2022-12-31 23:59:59
}

type OyCreatePaymentLinkResponse struct {
	PaymentLinkID string `json:"payment_link_id"`
	Message       string `json:"message"`
	EmailStatus   string `json:"email_status"`
	URL           string `json:"url"`
	Status        bool   `json:"status"`
}

type OyPaymentResponse struct {
	Success    bool        `json:"success"`
	Error      interface{} `json:"error"`
	Data       PaymentData `json:"data"`
	Reason     interface{} `json:"reason"`
	StatusCode int         `json:"status_code"`
}

type PaymentData struct {
	PartnerTxID         string  `json:"partner_tx_id"`
	TxRefNumber         string  `json:"tx_ref_number"`
	Amount              float64 `json:"amount"`
	SenderName          string  `json:"sender_name"`
	SenderPhone         string  `json:"sender_phone"`
	SenderNote          string  `json:"sender_note"`
	Status              string  `json:"status"`
	SenderBank          string  `json:"sender_bank"`
	IsInvoice           bool    `json:"is_invoice"`
	PaidAmount          float64 `json:"paid_amount"`
	PaymentMethod       string  `json:"payment_method"`
	Description         string  `json:"description"`
	Email               string  `json:"email"`
	PaymentReceivedTime string  `json:"payment_received_time"`
	SettlementTime      string  `json:"settlement_time"`
	SettlementStatus    string  `json:"settlement_status"`
	SettlementType      string  `json:"settlement_type"`
	Created             string  `json:"created"`
	Updated             string  `json:"updated"`
	Expiration          string  `json:"expiration"`
}

type OyCallback struct {
	PartnerTxID         string  `json:"partner_tx_id"`
	TxRefNumber         string  `json:"tx_ref_number"`
	Amount              float64 `json:"amount"`
	SenderName          string  `json:"sender_name"`
	SenderPhone         string  `json:"sender_phone"`
	SenderNote          string  `json:"sender_note"`
	Status              string  `json:"status"`
	SenderBank          string  `json:"sender_bank"`
	IsInvoice           bool    `json:"is_invoice"`
	PaidAmount          float64 `json:"paid_amount"`
	PaymentMethod       string  `json:"payment_method"`
	Description         string  `json:"description"`
	Email               string  `json:"email"`
	PaymentReceivedTime string  `json:"payment_received_time"`
	SettlementTime      string  `json:"settlement_time"`
	SettlementStatus    string  `json:"settlement_status"`
	SettlementType      string  `json:"settlement_type"`
	Created             string  `json:"created"`
	Updated             string  `json:"updated"`
	Expiration          string  `json:"expiration"`
}
