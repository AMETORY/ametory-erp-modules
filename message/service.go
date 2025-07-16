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

// NewMessageService creates a new instance of MessageService with the given database connection and context.
//
// It initializes the InboxService and ChatService and calls the Migrate method of the MessageService to create the necessary database schema.
func NewMessageService(ctx *context.ERPContext) *MessageService {
	service := MessageService{
		ctx:          ctx,
		InboxService: inbox.NewInboxService(ctx.DB, ctx),
		ChatService:  chat.NewChatService(ctx.DB, ctx),
	}
	service.Migrate()
	return &service
}

// Migrate migrates the database schema for the MessageService.
//
// If the SkipMigration flag is set to true in the context, this method
// will not perform any migration and will return nil. Otherwise, it will
// attempt to auto-migrate the database to include the InboxModel,
// InboxMessageModel, ChatChannelModel, and ChatMessageModel schemas.
// If the migration process encounters an error, it will return that error.
// Otherwise, it will return nil upon successful migration.
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
