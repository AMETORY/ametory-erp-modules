package whatsmeow_client

import (
	"github.com/AMETORY/ametory-erp-modules/shared/objects"
)

type WaMessage struct {
	JID             string                   `json:"jid"`
	Text            string                   `json:"text"`
	FileType        string                   `json:"file_type"`
	FileUrl         string                   `json:"file_url"`
	To              string                   `json:"to"`
	IsGroup         bool                     `json:"is_group"`
	RefID           *string                  `json:"ref_id"`
	RefFrom         *string                  `json:"ref_from"`
	RefText         *string                  `json:"ref_text"`
	ChatPresence    string                   `json:"chat_presence"`
	EventMessage    *objects.EventMessage    `json:"event_message,omitempty"`
	LocationMessage *objects.LocationMessage `json:"location_message,omitempty"`
	ContactMessage  *objects.ContactMessage  `json:"contact_message,omitempty"`
}
