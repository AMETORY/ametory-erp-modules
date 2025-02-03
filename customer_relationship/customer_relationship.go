package customer_relationship

import (
	"fmt"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/customer_relationship/whatsapp"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
)

type CustomerRelationshipService struct {
	ctx             *context.ERPContext
	whatsappService *whatsapp.WhatsappService
}

func NewCustomerRelationshipService(ctx *context.ERPContext) *CustomerRelationshipService {
	csService := CustomerRelationshipService{
		ctx:             ctx,
		whatsappService: whatsapp.NewWhatsappService(ctx.DB, ctx),
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
	return cs.ctx.DB.AutoMigrate(&models.WhatsappMessageModel{})
}
