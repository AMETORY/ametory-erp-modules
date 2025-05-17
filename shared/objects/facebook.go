package objects

type FacebookWebhookSender struct {
	ID string `json:"id"`
}

type FacebookWebhookRecipient struct {
	ID string `json:"id"`
}

type FacebookWebhookMessage struct {
	MID         string `json:"mid,omitempty"`
	Text        string `json:"text,omitempty"`
	Attachments []struct {
		Type    string `json:"type,omitempty"`
		Payload struct {
			URL string `json:"url,omitempty"`
		} `json:"payload,omitempty"`
	} `json:"attachments,omitempty"`
}

type FacebookWebhookMessaging struct {
	Sender    FacebookWebhookSender    `json:"sender"`
	Recipient FacebookWebhookRecipient `json:"recipient"`
	Timestamp int64                    `json:"timestamp"`
	Message   FacebookWebhookMessage   `json:"message"`
}

type FacebookWebhookEntry struct {
	Time      int64                      `json:"time"`
	ID        string                     `json:"id"`
	Messaging []FacebookWebhookMessaging `json:"messaging"`
}

type FacebookWebhookResponse struct {
	Object string                 `json:"object"`
	Entry  []FacebookWebhookEntry `json:"entry"`
}

type FacebookAccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type FacebookUser struct {
	Name              string `json:"name"`
	ID                string `json:"id"`
	ProfilePictureURL string `json:"profile_picture_url"`
	UserID            string `json:"user_id"`
}
