package models

type WhatsappList struct {
	Type                         string              `json:"type,omitempty" bson:"type,omitempty"`
	Header                       *WhatsappListHeader `json:"header,omitempty" bson:"header,omitempty"`
	Body                         WhatsappListBody    `json:"body,omitempty" bson:"body,omitempty"`
	Footer                       *WhatsappListFooter `json:"footer,omitempty" bson:"footer,omitempty"`
	Action                       WhatsappListAction  `json:"action,omitempty" bson:"action,omitempty"`
	WhatsappInteractiveMessageID *string             `json:"whatsapp_interactive_message_id,omitempty" bson:"whatsappInteractiveMessageId,omitempty"`
}

type WhatsappListHeader struct {
	Type string `json:"type,omitempty" bson:"type,omitempty"`
	Text string `json:"text,omitempty" bson:"text,omitempty"`
}
type WhatsappListBody struct {
	Text string `json:"text,omitempty" bson:"text,omitempty"`
}
type WhatsappListFooter struct {
	Text string `json:"text,omitempty" bson:"text,omitempty"`
}
type WhatsappListRows struct {
	ID          string `json:"id,omitempty" bson:"id,omitempty"`
	Title       string `json:"title,omitempty" bson:"title,omitempty"`
	Description string `json:"description,omitempty" bson:"description,omitempty"`
}
type WhatsappListSections struct {
	Title string             `json:"title,omitempty" bson:"title,omitempty"`
	Rows  []WhatsappListRows `json:"rows,omitempty" bson:"rows,omitempty"`
}
type WhatsappListAction struct {
	Button   string                 `json:"button,omitempty" bson:"button,omitempty"`
	Sections []WhatsappListSections `json:"sections,omitempty" bson:"sections,omitempty"`
}
