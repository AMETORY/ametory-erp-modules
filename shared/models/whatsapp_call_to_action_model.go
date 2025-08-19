package models

type WhatsappCallToAction struct {
	Type                         string                      `json:"type" bson:"type"`
	Header                       *WhatsappCallToActionHeader `json:"header,omitempty" bson:"header,omitempty"`
	Body                         WhatsappCallToActionBody    `json:"body" bson:"body"`
	Action                       WhatsappCallToActionAction  `json:"action" bson:"action"`
	Footer                       *WhatsappCallToActionFooter `json:"footer,omitempty" bson:"footer,omitempty"`
	WhatsappInteractiveMessageID *string                     `json:"whatsapp_interactive_message_id,omitempty" bson:"whatsappInteractiveMessageId,omitempty"`
}

type WhatsappCallToActionHeader struct {
	Type     string                              `json:"type" bson:"type"`
	Text     *string                             `json:"text,omitempty" bson:"text,omitempty"`
	Image    *WhatsappCallToActionHeaderImage    `json:"image,omitempty" bson:"image,omitempty"`
	Video    *WhatsappCallToActionHeaderVideo    `json:"video,omitempty" bson:"video,omitempty"`
	Document *WhatsappCallToActionHeaderDocument `json:"document,omitempty" bson:"document,omitempty"`
}

type WhatsappCallToActionHeaderImage struct {
	Link string `json:"link" bson:"link"`
}
type WhatsappCallToActionHeaderVideo struct {
	Link string `json:"link" bson:"link"`
}
type WhatsappCallToActionHeaderDocument struct {
	Link string `json:"link" bson:"link"`
}
type WhatsappCallToActionBody struct {
	Text string `json:"text" bson:"text"`
}
type WhatsappCallToActionFooter struct {
	Text string `json:"text" bson:"text"`
}

type WhatsappCallToActionAction struct {
	Name       string                         `json:"name" bson:"name"`
	Parameters WhatsappCallToActionParameters `json:"parameters" bson:"parameters"`
}

type WhatsappCallToActionParameters struct {
	DisplayText string `json:"display_text" bson:"displayText"`
	URL         string `json:"url" bson:"url"`
}
