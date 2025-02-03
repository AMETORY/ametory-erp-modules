package models

import (
	"errors"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WhatsappMessageModel struct {
	shared.BaseModel
	Sender    string        `gorm:"type:varchar(255)" json:"sender"`
	Receiver  string        `gorm:"type:varchar(255)" json:"receiver"`
	Message   string        `json:"message"`
	MediaURL  string        `gorm:"type:varchar(255)" json:"media_url"`
	MimeType  string        `gorm:"type:varchar(255)" json:"mime_type"`
	Session   string        `gorm:"type:varchar(255)" json:"session"`
	Info      string        `gorm:"type:json" json:"info"`
	ContactID *string       `json:"contact_id,omitempty" gorm:"column:contact_id"`
	Contact   *ContactModel `gorm:"foreignKey:ContactID" json:"contact,omitempty"`
}

func (m *WhatsappMessageModel) TableName() string {
	return "whatsapp_messages"
}

func (m *WhatsappMessageModel) BeforeCreate(tx *gorm.DB) error {
	if m.Session == "" {
		return errors.New("session is required")
	}
	if m.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	if m.Info == "" {
		tx.Statement.SetColumn("info", "{}")
	}

	var contact ContactModel
	err := tx.Select("phone", "id").First(&contact, "phone = ?", m.Sender).Error
	if err == nil {
		tx.Statement.SetColumn("contact_id", contact.ID)
	}

	return nil
}
