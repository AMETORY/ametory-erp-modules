package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"gorm.io/gorm"
)

type InstagramMessage struct {
	shared.BaseModel
	Message                   string                   `json:"message"`
	MediaURL                  string                   `gorm:"type:varchar(255)" json:"media_url"`
	MimeType                  string                   `gorm:"type:varchar(255)" json:"mime_type"`
	Session                   string                   `gorm:"type:varchar(255)" json:"session"`
	ContactID                 *string                  `json:"contact_id,omitempty" gorm:"column:contact_id"`
	Contact                   *ContactModel            `gorm:"foreignKey:ContactID" json:"contact,omitempty"`
	CompanyID                 *string                  `json:"company_id,omitempty" gorm:"column:company_id"`
	Company                   *CompanyModel            `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	IsFromMe                  bool                     `json:"is_from_me"`
	IsGroup                   bool                     `json:"is_group"`
	IsReplied                 bool                     `json:"is_replied" gorm:"default:false"`
	SentAt                    *time.Time               `json:"sent_at" gorm:"-"`
	IsRead                    bool                     `json:"is_read" gorm:"default:false"`
	MessageID                 *string                  `json:"message_id" gorm:"column:message_id"`
	ResponseTime              *float64                 `json:"response_time"`
	MemberID                  *string                  `json:"member_id,omitempty" gorm:"column:member_id"`
	Member                    *MemberModel             `gorm:"foreignKey:MemberID" json:"member,omitempty"`
	UserID                    *string                  `json:"user_id,omitempty" gorm:"column:user_id"`
	User                      *UserModel               `gorm:"foreignKey:UserID" json:"user,omitempty"`
	IsNew                     bool                     `json:"is_new" gorm:"default:false"`
	RefID                     *string                  `json:"ref_id,omitempty" gorm:"column:ref_id"`
	IsAutoPilot               bool                     `json:"is_auto_pilot" gorm:"default:false"`
	InstagramMessageSessionID *string                  `json:"telegram_message_session_id,omitempty" gorm:"column:telegram_message_session_id"`
	InstagramMessageSession   *InstagramMessageSession `gorm:"foreignKey:InstagramMessageSessionID" json:"telegram_message_session,omitempty"`
}

func (t *InstagramMessage) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = utils.Uuid()
	}
	return nil
}

type InstagramMessageSession struct {
	shared.BaseModel
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
	Ref          any           `json:"ref,omitempty" gorm:"-"`
	IsHumanAgent bool          `json:"is_human_agent"`
	CountUnread  int           `json:"count_unread" gorm:"-"`
}

func (t *InstagramMessageSession) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = utils.Uuid()
	}
	return nil
}
