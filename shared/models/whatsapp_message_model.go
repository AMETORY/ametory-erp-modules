package models

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WhatsappMessageModel struct {
	shared.BaseModel
	JID         string                 `gorm:"type:varchar(255)" json:"jid"`
	Sender      string                 `gorm:"type:varchar(255)" json:"sender"`
	Receiver    string                 `gorm:"type:varchar(255)" json:"receiver"`
	Message     string                 `json:"message"`
	MediaURL    string                 `gorm:"type:varchar(255)" json:"media_url"`
	MimeType    string                 `gorm:"type:varchar(255)" json:"mime_type"`
	Session     string                 `gorm:"type:varchar(255)" json:"session"`
	Info        string                 `gorm:"type:json" json:"-"`
	MessageInfo map[string]interface{} `gorm:"-" json:"message_info"`
	ContactID   *string                `json:"contact_id,omitempty" gorm:"column:contact_id"`
	Contact     *ContactModel          `gorm:"foreignKey:ContactID" json:"contact,omitempty"`
	CompanyID   *string                `json:"company_id,omitempty" gorm:"column:company_id"`
	Company     *CompanyModel          `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	IsFromMe    bool                   `json:"is_from_me"`
	IsGroup     bool                   `json:"is_group"`
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

func (m *WhatsappMessageModel) AfterFind(tx *gorm.DB) error {
	if m.Info != "" {
		var info map[string]interface{}
		err := json.Unmarshal([]byte(m.Info), &info)
		if err != nil {
			return err
		}
		m.MessageInfo = info
	}
	return nil
}

type WhatsappMessageSession struct {
	shared.BaseModel
	JID          string        `gorm:"type:varchar(255)" json:"jid"`
	Session      string        `gorm:"type:varchar(255)" json:"session"`
	SessionName  string        `gorm:"type:varchar(255)" json:"session_name"`
	LastOnlineAt *time.Time    `json:"last_online_at"`
	LastMessage  string        `json:"last_message"`
	CompanyID    *string       `json:"company_id,omitempty" gorm:"column:company_id"`
	Company      *CompanyModel `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
}

func (m *WhatsappMessageSession) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return nil
}
