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

// SendCustomerServiceMessage sends a customer service message using the input provided to the service.
//
// It calls SendCSMessage on the input CustomerServiceMessage to send the message and processes the response to
// extract the message data, which is stored in the input data. If the input specifies that the message should be saved,
// it calls SaveMessage to save the message data.
//
// Returns the message data if successful, or an error if any step fails.
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

// NewCustomerRelationshipService creates a new instance of CustomerRelationshipService.
//
// It initializes the service with the given ERP context and sets up the Whatsapp,
// Form, Telegram, and Instagram services. It also performs migration for the necessary
// database models. If the migration fails, it logs the error.
//
// Returns a pointer to the newly created CustomerRelationshipService.

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

// Migrate migrates the database schema for the CustomerRelationshipService.
//
// If the SkipMigration flag is set to true in the context, this method
// will not perform any migration and will return nil. Otherwise, it will
// attempt to auto-migrate the database to include the schemas for
// WhatsappMessageModel, WhatsappMessageReaction, WhatsappMessageSession,
// FormTemplate, FormModel, FormResponseModel, WhatsappMessageTemplate,
// MessageTemplate, TelegramMessage, TelegramMessageSession,
// InstagramMessage, and InstagramMessageSession.
// If the migration process encounters an error, it will return that error.
// Otherwise, it will return nil upon successful migration.

func (cs *CustomerRelationshipService) Migrate() error {
	if cs.ctx.SkipMigration {
		return nil
	}

	return cs.ctx.DB.AutoMigrate(
		&models.WhatsappMessageModel{},
		&models.WhatsappMessageReaction{},
		&models.WhatsappMessageSession{},
		&models.WhatsappInteractiveMessage{},
		&models.FormTemplate{},
		&models.FormModel{},
		&models.FormResponseModel{},
		&models.WhatsappMessageTemplate{},
		&models.MessageTemplate{},
		&models.TelegramMessage{},
		&models.TelegramMessageSession{},
		&models.TiktokMessageSession{},
		&models.InstagramMessage{},
		&models.InstagramMessageSession{},
	)
}
