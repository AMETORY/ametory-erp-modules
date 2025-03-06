package message

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/message/inbox"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"gorm.io/gorm"
)

type MessageService struct {
	db           *gorm.DB
	ctx          *context.ERPContext
	InboxService *inbox.InboxService
}

func NewMessageService(db *gorm.DB, ctx *context.ERPContext) *MessageService {
	return &MessageService{
		db:           db,
		ctx:          ctx,
		InboxService: inbox.NewInboxService(db, ctx),
	}
}

func (cs *MessageService) Migrate() error {
	if cs.ctx.SkipMigration {
		return nil
	}
	return cs.ctx.DB.AutoMigrate(
		&models.InboxModel{},
		&models.InboxMessageModel{},
	)
}
