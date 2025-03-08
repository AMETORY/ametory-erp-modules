package message

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/message/chat"
	"github.com/AMETORY/ametory-erp-modules/message/inbox"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
)

type MessageService struct {
	ctx          *context.ERPContext
	InboxService *inbox.InboxService
	ChatService  *chat.ChatService
}

func NewMessageService(ctx *context.ERPContext) *MessageService {
	service := MessageService{
		ctx:          ctx,
		InboxService: inbox.NewInboxService(ctx.DB, ctx),
	}
	service.Migrate()
	return &service
}

func (cs *MessageService) Migrate() error {
	if cs.ctx.SkipMigration {
		return nil
	}
	return cs.ctx.DB.AutoMigrate(
		&models.InboxModel{},
		&models.InboxMessageModel{},
		&models.ChatChannelModel{},
		&models.ChatMessageModel{},
	)
}
