package models

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WhatsappMessageModel struct {
	shared.BaseModel
	JID          string                 `gorm:"type:varchar(255)" json:"jid"`
	Sender       string                 `gorm:"type:varchar(255)" json:"sender"`
	Receiver     string                 `gorm:"type:varchar(255)" json:"receiver"`
	Message      string                 `json:"message"`
	MediaURL     string                 `gorm:"type:varchar(255)" json:"media_url"`
	MimeType     string                 `gorm:"type:varchar(255)" json:"mime_type"`
	Session      string                 `gorm:"type:varchar(255)" json:"session"`
	Info         string                 `gorm:"type:json" json:"-"`
	MessageInfo  map[string]interface{} `gorm:"-" json:"message_info"`
	ContactID    *string                `json:"contact_id,omitempty" gorm:"column:contact_id"`
	Contact      *ContactModel          `gorm:"foreignKey:ContactID" json:"contact,omitempty"`
	CompanyID    *string                `json:"company_id,omitempty" gorm:"column:company_id"`
	Company      *CompanyModel          `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	IsFromMe     bool                   `json:"is_from_me"`
	IsGroup      bool                   `json:"is_group"`
	IsReplied    bool                   `json:"is_replied" gorm:"default:false"`
	SentAt       *time.Time             `json:"sent_at" gorm:"-"`
	IsRead       bool                   `json:"is_read" gorm:"default:false"`
	MessageID    *string                `json:"message_id" gorm:"column:message_id"`
	ResponseTime *float64               `json:"response_time"`
	MemberID     *string                `json:"member_id,omitempty" gorm:"column:member_id"`
	Member       *MemberModel           `gorm:"foreignKey:MemberID" json:"member,omitempty"`
	UserID       *string                `json:"user_id,omitempty" gorm:"column:user_id"`
	User         *UserModel             `gorm:"foreignKey:UserID" json:"user,omitempty"`
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

	// var contact ContactModel
	// err := tx.Select("phone", "id").First(&contact, "phone = ?", m.Sender).Error
	// if err == nil {
	// 	tx.Statement.SetColumn("contact_id", contact.ID)
	// }

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
		sentAt, ok := info["Timestamp"].(string)
		if ok {
			t, err := time.Parse(time.RFC3339, sentAt)
			if err == nil {
				m.SentAt = &t
			}
		}

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
	ContactID    *string       `json:"contact_id,omitempty" gorm:"column:contact_id"`
	Contact      *ContactModel `gorm:"foreignKey:ContactID" json:"contact,omitempty"`
	RefID        *string       `json:"ref_id,omitempty"`
	RefType      *string       `json:"ref_type,omitempty"`
	Ref          interface{}   `json:"ref,omitempty" gorm:"-"`
	IsHumanAgent bool          `json:"is_human_agent"`
	CountUnread  int           `json:"count_unread" gorm:"-"`
}

func (m *WhatsappMessageSession) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return nil
}

type WhatsappMessageTemplate struct {
	shared.BaseModel
	Title       string            `gorm:"type:varchar(255)" json:"title"`
	ShortCut    string            `gorm:"type:varchar(255)" json:"short_cut"`
	Description string            `gorm:"type:text" json:"description"`
	CompanyID   *string           `json:"company_id,omitempty" gorm:"column:company_id;constraint:OnDelete:CASCADE;"`
	Company     *CompanyModel     `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	UserID      *string           `json:"user_id,omitempty" gorm:"column:user_id;constraint:OnDelete:CASCADE;"`
	User        *UserModel        `gorm:"foreignKey:UserID" json:"user,omitempty"`
	MemberID    *string           `json:"member_id,omitempty" gorm:"column:member_id;constraint:OnDelete:CASCADE;"`
	Member      *MemberModel      `gorm:"foreignKey:MemberID" json:"member,omitempty"`
	Messages    []MessageTemplate `gorm:"-" json:"messages,omitempty"`
}

func (m *WhatsappMessageTemplate) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}

	if m.ShortCut == "" {
		m.ShortCut = utils.URLify(m.Title)
	}
	return nil
}

func (m *WhatsappMessageTemplate) AfterFind(tx *gorm.DB) error {
	var messages []MessageTemplate
	tx.Where("whatsapp_message_template_id = ?", m.ID).Preload("Products").Find(&messages)
	m.Messages = messages
	return nil
}

type MessageTemplate struct {
	shared.BaseModel
	WhatsappMessageTemplateID *string                  `json:"whatsapp_message_template_id,omitempty" gorm:"column:whatsapp_message_template_id;constraint:OnDelete:CASCADE;"`
	WhatsappMessageTemplate   *WhatsappMessageTemplate `gorm:"foreignKey:WhatsappMessageTemplateID" json:"whatsapp_message_template,omitempty"`
	Type                      string                   `json:"type"`
	Header                    string                   `json:"header"`
	Body                      string                   `json:"body"`
	Footer                    string                   `json:"footer"`
	ButtonText                string                   `json:"button_text"`
	ButtonUrl                 string                   `json:"button_url"`
	Files                     []FileModel              `json:"files,omitempty" gorm:"-"`
	Products                  []ProductModel           `gorm:"many2many:whatsapp_message_template_products" json:"products,omitempty"`
}

func (m *MessageTemplate) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return nil
}

func (m *MessageTemplate) AfterFind(tx *gorm.DB) error {
	tx.Model(&FileModel{}).Where("ref_id = ? AND ref_type = ?", m.ID, "message_template").Find(&m.Files)
	return nil
}
