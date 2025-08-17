package objects

type WebhookEntryChange struct {
	Field string                  `json:"field"`
	Value WebhookEntryChangeValue `json:"value"`
}

type WebhookEntryChangeValue struct {
	Contacts         []WebhookEntryChangeContact `json:"contacts"`
	Messages         []WebhookEntryChangeMessage `json:"messages"`
	MessagingProduct string                      `json:"messaging_product"`
	Metadata         *WebhookEntryChangeMetadata `json:"metadata"`
	From             *WebhookEntryChangeFrom     `json:"from"`
	ID               string                      `json:"id"`
	Media            *WebhookEntryChangeMedia    `json:"media"`
	Text             string                      `json:"text"`
}

type WebhookEntryChangeContext struct {
	From *WebhookEntryChangeFrom `json:"from"`
	ID   string                  `json:"id"`
}
type WebhookImage struct {
	Caption  string `json:"caption"`
	ID       string `json:"id"`
	MimeType string `json:"mime_type"`
	Sha256   string `json:"sha256"`
}

type WebhookEntryChangeFrom struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

type WebhookEntryChangeMedia struct {
	ID               string `json:"id"`
	MediaProductType string `json:"media_product_type"`
}

type WebhookEntryChangeContact struct {
	Profile struct {
		Name string `json:"name"`
	} `json:"profile"`
	WAID string `json:"wa_id"`
}

type WebhookEntryChangeMessage struct {
	Context *WebhookEntryChangeContext `json:"context"`
	From    string                     `json:"from"`
	ID      string                     `json:"id"`
	Text    struct {
		Body string `json:"body"`
	} `json:"text"`
	Timestamp string        `json:"timestamp"`
	Type      string        `json:"type"`
	Image     *WebhookImage `json:"image,omitempty"`
}

type WebhookEntryChangeMetadata struct {
	DisplayPhoneNumber string `json:"display_phone_number"`
	PhoneNumberID      string `json:"phone_number_id"`
}

type WebhookEntry struct {
	Changes                  []WebhookEntryChange       `json:"changes"`
	ID                       string                     `json:"id"`
	FacebookWebhookMessaging []FacebookWebhookMessaging `json:"messaging,omitempty"`
	Time                     int64                      `json:"time"`
}

type WhatsappApiWebhookRequest struct {
	Entry  []WebhookEntry `json:"entry"`
	Object string         `json:"object"`
}

type FacebookMedia struct {
	URL              string `json:"url"`
	MimeType         string `json:"mime_type"`
	SHA256           string `json:"sha256"`
	FileSize         int    `json:"file_size"`
	ID               string `json:"id"`
	MessagingProduct string `json:"messaging_product"`
}

type WaResponse struct {
	MessagingProduct string `json:"messaging_product"`
	Contacts         []struct {
		Input string `json:"input"`
		WaID  string `json:"wa_id"`
	} `json:"contacts"`
	Messages []struct {
		ID            string `json:"id"`
		MessageStatus string `json:"message_status"`
	} `json:"messages"`
}
