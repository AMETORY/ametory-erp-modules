package manufacture

import (
	"fmt"
	"log"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/inventory"
	"github.com/AMETORY/ametory-erp-modules/manufacture/bom"
	"github.com/AMETORY/ametory-erp-modules/manufacture/work_order"
)

type ManufactureService struct {
	ctx              *context.ERPContext
	inventoryService *inventory.InventoryService
}

// NewManufactureService creates a new instance of ManufactureService.
//
// It takes an ERPContext and an InventoryService as parameter and returns a pointer to a ManufactureService.
//
// The service will also call the Migrate() method after creation to migrate the database schema.
// If the migration fails, the service will return nil.
func NewManufactureService(ctx *context.ERPContext, inventoryService *inventory.InventoryService) *ManufactureService {
	fmt.Println("INIT MANUFACTURE SERVICE")
	service := &ManufactureService{
		ctx:              ctx,
		inventoryService: inventoryService,
	}
	err := service.Migrate()
	if err != nil {
		fmt.Println("INIT MANUFACTURE ERROR", err)
		return nil
	}
	return service

}

// Migrate runs the database migrations for the manufacture module.
//
// It first checks if the SkipMigration flag is set in the context. If it is, the
// function returns immediately.
//
// It then calls the Migrate functions of the BOM and Work Order services, passing
// the database connection from the context. If either of these calls returns an
// error, the function logs the error and returns it.
func (s *ManufactureService) Migrate() error {
	if s.ctx.SkipMigration {
		return nil
	}
	if err := bom.Migrate(s.ctx.DB); err != nil {
		log.Println("ERROR BOM", err)
		return err
	}
	if err := work_order.Migrate(s.ctx.DB); err != nil {
		log.Println("ERROR WORK ORDER", err)
		return err
	}

	return nil
}
