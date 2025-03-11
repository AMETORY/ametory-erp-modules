package customer_relationship

import (
	"fmt"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/customer_relationship/form"
	"github.com/AMETORY/ametory-erp-modules/customer_relationship/whatsapp"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
)

type CustomerRelationshipService struct {
	ctx             *context.ERPContext
	WhatsappService *whatsapp.WhatsappService
	FormService     *form.FormService
}

func NewCustomerRelationshipService(ctx *context.ERPContext) *CustomerRelationshipService {
	csService := CustomerRelationshipService{
		ctx:             ctx,
		WhatsappService: whatsapp.NewWhatsappService(ctx.DB, ctx),
		FormService:     form.NewFormService(ctx.DB, ctx),
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
		&models.FormTemplate{},
		&models.FormModel{},
		&models.FormResponseModel{},
	)
}
