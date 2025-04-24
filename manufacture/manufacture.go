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
