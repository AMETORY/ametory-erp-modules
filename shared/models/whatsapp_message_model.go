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
	shared.BaseModel         `bson:"base"`
	JID                      string                    `gorm:"type:varchar(255)" json:"jid" bson:"jid"`
	Sender                   string                    `gorm:"type:varchar(255)" json:"sender" bson:"sender"`
	Receiver                 string                    `gorm:"type:varchar(255)" json:"receiver" bson:"receiver"`
	Message                  string                    `json:"message" bson:"message"`
	QuotedMessage            *string                   `json:"quotedMessage" bson:"quotedMessage"`
	QuotedMessageID          *string                   `json:"quotedMessageId" bson:"quotedMessageId"`
	MediaURL                 string                    `gorm:"type:varchar(255)" json:"mediaUrl" bson:"mediaUrl"`
	MimeType                 string                    `gorm:"type:varchar(255)" json:"mimeType" bson:"mimeType"`
	Session                  string                    `gorm:"type:varchar(255)" json:"session" bson:"session"`
	Info                     string                    `gorm:"type:json" json:"-" bson:"info"`
	MessageInfo              map[string]interface{}    `gorm:"-" json:"messageInfo" bson:"messageInfo"`
	ContactID                *string                   `json:"contactId,omitempty" gorm:"column:contact_id" bson:"contactId"`
	Contact                  *ContactModel             `gorm:"foreignKey:ContactID" json:"contact,omitempty" bson:"contact"`
	CompanyID                *string                   `json:"companyId,omitempty" gorm:"column:company_id" bson:"companyId"`
	Company                  *CompanyModel             `gorm:"foreignKey:CompanyID" json:"company,omitempty" bson:"company"`
	IsFromMe                 bool                      `json:"isFromMe" bson:"isFromMe"`
	IsGroup                  bool                      `json:"isGroup" bson:"isGroup"`
	IsReplied                bool                      `json:"isReplied" gorm:"default:false" bson:"isReplied"`
	SentAt                   *time.Time                `json:"sentAt" gorm:"-" bson:"sentAt"`
	IsRead                   bool                      `json:"isRead" gorm:"default:false" bson:"isRead"`
	MessageID                *string                   `json:"messageId" gorm:"column:message_id" bson:"messageId"`
	ResponseTime             *float64                  `json:"responseTime" bson:"responseTime"`
	MemberID                 *string                   `json:"memberId,omitempty" gorm:"column:member_id" bson:"memberId"`
	Member                   *MemberModel              `gorm:"foreignKey:MemberID" json:"member,omitempty" bson:"member"`
	UserID                   *string                   `json:"userId,omitempty" gorm:"column:user_id" bson:"userId"`
	User                     *UserModel                `gorm:"foreignKey:UserID" json:"user,omitempty" bson:"user"`
	IsNew                    bool                      `json:"isNew" gorm:"default:false" bson:"isNew"`
	RefID                    *string                   `json:"refId,omitempty" gorm:"column:ref_id" bson:"refId"`
	IsAutoPilot              bool                      `json:"isAutoPilot" gorm:"default:false" bson:"isAutoPilot"`
	WhatsappMessageReactions []WhatsappMessageReaction `gorm:"foreignKey:WhatsappMessageID" json:"whatsappMessageReactions,omitempty" bson:"whatsappMessageReactions"`
}

type WhatsappMessageReaction struct {
	shared.BaseModel
	Reaction          string                `gorm:"type:varchar(255)" json:"reaction"`
	WhatsappMessageID *string               `json:"whatsapp_message_id,omitempty" gorm:"column:whatsapp_message_id;uniqueIndex:idx_whatsapp_message_reactions_whatsapp_message_id_contact_id"`
	WhatsappMessage   *WhatsappMessageModel `gorm:"foreignKey:WhatsappMessageID" json:"whatsapp_message,omitempty"`
	Contact           *ContactModel         `gorm:"foreignKey:ContactID" json:"contact,omitempty"`
	ContactID         *string               `json:"contact_id,omitempty" gorm:"column:contact_id;uniqueIndex:idx_whatsapp_message_reactions_whatsapp_message_id_contact_id"`
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
	shared.BaseModel `bson:"base"`
	JID              string        `gorm:"type:varchar(255);index" json:"jid" bson:"jid"`
	Session          string        `gorm:"type:varchar(255);index" json:"session" bson:"session"`
	SessionName      string        `gorm:"type:varchar(255)" json:"sessionName" bson:"sessionName"`
	LastOnlineAt     *time.Time    `json:"lastOnlineAt" bson:"lastOnlineAt"`
	LastMessage      string        `json:"lastMessage" bson:"lastMessage"`
	CompanyID        *string       `json:"companyID,omitempty" gorm:"column:company_id" bson:"companyId"`
	Company          *CompanyModel `gorm:"foreignKey:CompanyID" json:"company,omitempty" bson:"company"`
	ContactID        *string       `json:"contactID,omitempty" gorm:"column:contact_id" bson:"contactId"`
	Contact          *ContactModel `gorm:"foreignKey:ContactID" json:"contact,omitempty" bson:"contact"`
	RefID            *string       `json:"refID,omitempty" gorm:"index" bson:"refId"`
	RefType          *string       `json:"refType,omitempty" bson:"refType"`
	Ref              any           `json:"ref,omitempty" gorm:"-" bson:"-"`
	IsHumanAgent     bool          `json:"isHumanAgent" bson:"isHumanAgent"`
	IsGroup          bool          `json:"isGroup" gorm:"default:false" bson:"isGroup"`
	CountUnread      int           `json:"countUnread" gorm:"-" bson:"-"`
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
