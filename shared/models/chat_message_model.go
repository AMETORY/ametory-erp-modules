package models

import (
	"fmt"
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ChatChannelModel struct {
	shared.BaseModel
	Name               string         `gorm:"type:varchar(255)" json:"name"`
	Description        string         `gorm:"type:text" json:"description"`
	Icon               string         `gorm:"type:varchar(255)" json:"icon"`
	Color              string         `gorm:"type:varchar(255)" json:"color"`
	ParticipantUsers   []*UserModel   `gorm:"many2many:chat_channel_participant_users;constraint:OnDelete:CASCADE;" json:"participant_users,omitempty"`
	ParticipantMembers []*MemberModel `gorm:"many2many:chat_channel_participant_members;constraint:OnDelete:CASCADE;" json:"participant_members,omitempty"`
	CreatedBy          *UserModel     `gorm:"foreignKey:CreatedByUserID;constraint:OnDelete:SET NULL;" json:"created_by,omitempty"`
	CreatedByUserID    *string        `gorm:"type:char(36);index" json:"created_by_user_id,omitempty"`
	CreatedByMember    *MemberModel   `gorm:"foreignKey:CreatedByMemberID;constraint:OnDelete:SET NULL;" json:"created_by_member,omitempty"`
	CreatedByMemberID  *string        `gorm:"type:char(36);index" json:"created_by_member_id,omitempty"`
	Avatar             *FileModel     `gorm:"-" json:"avatar,omitempty"`
	RefID              string         `gorm:"type:char(36);index" json:"ref_id,omitempty"`
	RefType            string         `gorm:"type:varchar(255);index" json:"ref_type,omitempty"`
}

func (ChatChannelModel) TableName() string {
	return "chat_channels"
}

func (channel *ChatChannelModel) BeforeCreate(tx *gorm.DB) (err error) {
	if channel.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	// Add any logic to be executed before creating a ChatChannelModel record
	return nil
}

func (channel *ChatChannelModel) AfterFind(tx *gorm.DB) error {
	var file FileModel
	fmt.Println("AFTER FIND", channel.ID)
	err := tx.Model(&file).
		Where("ref_id = ? and ref_type = ?", channel.ID, "chat").
		First(&file).Error
	if err == nil {
		channel.Avatar = &file
	}

	return nil
}

// ChatMessageModel adalah model database untuk menampung data chat message
type ChatMessageModel struct {
	shared.BaseModel
	ChatChannelID   *string            `gorm:"type:char(36);index" json:"chat_channel_id,omitempty"`
	ChatChannel     *ChatChannelModel  `gorm:"foreignKey:ChatChannelID;constraint:OnDelete:CASCADE;" json:"chat_channel,omitempty"`
	SenderUserID    *string            `gorm:"type:char(36);index" json:"sender_user_id,omitempty"`
	SenderUser      *UserModel         `gorm:"foreignKey:SenderUserID;constraint:OnDelete:CASCADE;" json:"sender_user,omitempty"`
	SenderMemberID  *string            `gorm:"type:char(36);index" json:"sender_member_id,omitempty"`
	SenderMember    *MemberModel       `gorm:"foreignKey:SenderMemberID;constraint:OnDelete:CASCADE;" json:"sender_member,omitempty"`
	Message         string             `gorm:"type:text" json:"message,omitempty"`
	Type            string             `gorm:"type:varchar(255);default:CHAT" json:"type,omitempty"`
	Files           []FileModel        `gorm:"-" json:"files,omitempty"`
	Date            *time.Time         `json:"date"`
	ReadedBy        []*UserModel       `gorm:"many2many:chat_message_read_by_users;constraint:OnDelete:CASCADE;" json:"read_by,omitempty"`
	ReadedByMembers []*MemberModel     `gorm:"many2many:chat_message_read_by_members;constraint:OnDelete:CASCADE;" json:"read_by_members,omitempty"`
	ParentID        *string            `gorm:"type:char(36);index" json:"parent_id,omitempty"`
	ParentMessage   *ChatMessageModel  `gorm:"foreignKey:ParentID;constraint:OnDelete:CASCADE;" json:"parent_message,omitempty"`
	Replies         []ChatMessageModel `gorm:"-" json:"replies,omitempty"`
	ChatData        string             `gorm:"type:JSON" json:"-"`
	Data            interface{}        `gorm:"-" json:"data,omitempty"`
	RepliesCount    int64              `gorm:"-" json:"replies_count,omitempty"`
	FilesCount      int64              `gorm:"-" json:"files_count,omitempty"`
}

func (ChatMessageModel) TableName() string {
	return "chat_messages"
}

func (message *ChatMessageModel) BeforeCreate(tx *gorm.DB) (err error) {
	if message.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	// Add any custom logic before creating a ChatMessageModel record
	return nil
}

func (message *ChatMessageModel) AfterFind(tx *gorm.DB) (err error) {
	if message.Date == nil {
		message.Date = message.CreatedAt
		tx.Save(message)
	}
	tx.Model(&ChatMessageModel{}).Where("parent_id = ?", message.ID).Count(&message.RepliesCount)
	tx.Model(&FileModel{}).Where("ref_id = ? AND ref_type = ?", message.ID, "chat").Count(&message.FilesCount)
	return
}

func (message *ChatMessageModel) GetFiles(tx *gorm.DB) error {
	var files []FileModel
	if err := tx.Where("ref_id = ? AND ref_type = ?", message.ID, "chat").Find(&files).Error; err != nil {
		return err
	}
	message.Files = files
	return nil
}
func (message *ChatMessageModel) GetReplies(tx *gorm.DB) error {
	var replies []ChatMessageModel
	if err := tx.Where("parent_id = ?", message.ID).Find(&replies).Error; err != nil {
		return err
	}
	message.Replies = replies
	return nil
}

// ChatMessageReadByMember adalah model database untuk menampung data read by member di chat message
type ChatMessageReadByMember struct {
	ChatMessageModelID string            `gorm:"type:char(36);index"`
	ChatMessageModel   *ChatMessageModel `gorm:"foreignKey:ChatMessageModelID;constraint:OnDelete:CASCADE;"`
	MemberModelID      string            `gorm:"type:char(36);index"`
	MemberModel        *MemberModel      `gorm:"foreignKey:MemberModelID;constraint:OnDelete:CASCADE;"`
}

func (ChatMessageReadByMember) TableName() string {
	return "chat_message_read_by_members"
}

// ChatMessageReadByUser adalah model database untuk menampung data read by user di chat message
type ChatMessageReadByUser struct {
	ChatMessageModelID string            `gorm:"type:char(36);index"`
	ChatMessageModel   *ChatMessageModel `gorm:"foreignKey:ChatMessageModelID;constraint:OnDelete:CASCADE;"`
	UserModelID        string            `gorm:"type:char(36);index"`
	UserModel          *UserModel        `gorm:"foreignKey:UserModelID;constraint:OnDelete:CASCADE;"`
}

func (ChatMessageReadByUser) TableName() string {
	return "chat_message_read_by_users"
}
