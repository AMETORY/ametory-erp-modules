package models

import (
	"encoding/json"
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TiktokMessageSession struct {
	shared.BaseModel
	Session      string           `gorm:"type:varchar(255)" json:"session"`
	SessionName  string           `gorm:"type:varchar(255)" json:"session_name"`
	CreateTime   *time.Time       `json:"create_time"`
	LastMessage  *json.RawMessage `json:"last_message"`
	LastMsgTime  *time.Time       `json:"last_msg_time"`
	CompanyID    *string          `json:"company_id,omitempty" gorm:"column:company_id"`
	Company      *CompanyModel    `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	Participant  *json.RawMessage `json:"participant,omitempty" gorm:"column:participant"`
	RefID        *string          `json:"ref_id,omitempty"`
	RefType      *string          `json:"ref_type,omitempty"`
	Ref          any              `json:"ref,omitempty" gorm:"-"`
	IsHumanAgent bool             `json:"is_human_agent"`
	CountUnread  int              `json:"count_unread" gorm:"-"`
}

// BeforeCreate is a GORM hook that will be triggered before a new record is created.
func (session *TiktokMessageSession) BeforeCreate(tx *gorm.DB) (err error) {
	if session.ID == "" {
		session.ID = uuid.New().String()
	}
	return
}

type TiktokMessage struct {
	Content        string `json:"content"`
	ConversationID string `json:"conversation_id"`
	CreateTime     *int64 `json:"create_time"`
	IsVisible      bool   `json:"is_visible"`
	MessageID      string `json:"message_id"`
	Index          string `json:"index"`
	Type           string `json:"type"`
	Sender         struct {
		ImUserID string `json:"im_user_id"`
		Role     string `json:"role"`
	} `json:"sender"`
}
