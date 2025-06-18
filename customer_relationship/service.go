package customer_relationship

import (
	"fmt"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/customer_relationship/form"
	"github.com/AMETORY/ametory-erp-modules/customer_relationship/instagram"
	"github.com/AMETORY/ametory-erp-modules/customer_relationship/telegram"
	"github.com/AMETORY/ametory-erp-modules/customer_relationship/whatsapp"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
)

type CustomerServiceMessage interface {
	SendCSMessage() (any, error)
}

func SendCustomerServiceMessage(cs CustomerServiceMessage) (any, error) {
	return cs.SendCSMessage()
}

type CustomerRelationshipService struct {
	ctx              *context.ERPContext
	WhatsappService  *whatsapp.WhatsappService
	FormService      *form.FormService
	TelegramService  *telegram.TelegramService
	InstagramService *instagram.InstagramService
}

func NewCustomerRelationshipService(ctx *context.ERPContext) *CustomerRelationshipService {
	csService := CustomerRelationshipService{
		ctx:              ctx,
		WhatsappService:  whatsapp.NewWhatsappService(ctx.DB, ctx),
		FormService:      form.NewFormService(ctx.DB, ctx),
		TelegramService:  telegram.NewTelegramService(ctx),
		InstagramService: instagram.NewInstagramService(ctx),
	}
	err := csService.Migrate()
	if err != nil {
		fmt.Println("ERROR MIGRATE", err)
	}
	return &csService
}

func (cs *CustomerRelationshipService) Migrate() error {
	if cs.ctx.SkipMigration {
		return nil
	}
	return cs.ctx.DB.AutoMigrate(
		&models.WhatsappMessageModel{},
		&models.WhatsappMessageReaction{},
		&models.WhatsappMessageSession{},
		&models.FormTemplate{},
		&models.FormModel{},
		&models.FormResponseModel{},
		&models.WhatsappMessageTemplate{},
		&models.MessageTemplate{},
		&models.TelegramMessage{},
		&models.TelegramMessageSession{},
		&models.InstagramMessage{},
		&models.InstagramMessageSession{},
	)
}
