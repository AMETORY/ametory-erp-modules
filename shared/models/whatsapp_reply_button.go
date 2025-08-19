package models

type WhatsappReplyButton struct {
	Type                         string                      `json:"type" bson:"type"`
	Header                       *WhatsappCallToActionHeader `json:"header,omitempty" bson:"header,omitempty"`
	Body                         WhatsappCallToActionBody    `json:"body" bson:"body"`
	Action                       WhatsappReplyButtonList     `json:"action" bson:"action"`
	Footer                       *WhatsappCallToActionFooter `json:"footer,omitempty" bson:"footer,omitempty"`
	WhatsappInteractiveMessageID *string                     `json:"whatsapp_interactive_message_id,omitempty" bson:"whatsappInteractiveMessageId,omitempty"`
}

type WhatsappReplyButtonList struct {
	Buttons []struct {
		Type  string `json:"type"`
		Reply struct {
			ID    string `json:"id"`
			Title string `json:"title"`
		} `json:"reply"`
	} `json:"buttons"`
}
