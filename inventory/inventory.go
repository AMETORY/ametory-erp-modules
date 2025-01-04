package inventory

import (
	"fmt"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/inventory/product"
	stockmovement "github.com/AMETORY/ametory-erp-modules/inventory/stock_movement"
	"github.com/AMETORY/ametory-erp-modules/inventory/warehouse"
	"gorm.io/gorm"
)

type InventoryService struct {
	ctx                    *context.ERPContext
	ProductService         *product.ProductService
	ProductCategoryService *product.ProductCategoryService
	WarehouseService       *warehouse.WarehouseService
	StockMovementService   *stockmovement.StockMovementService
}

func NewInventoryService(ctx *context.ERPContext) *InventoryService {
	fmt.Println("INIT INVENTORY SERVICE")
	var service = InventoryService{
		ctx:                    ctx,
		ProductService:         product.NewProductService(ctx.DB, ctx),
		ProductCategoryService: product.NewProductCategoryService(ctx.DB, ctx),
		WarehouseService:       warehouse.NewWarehouseService(ctx.DB, ctx),
		StockMovementService:   stockmovement.NewStockMovementService(ctx.DB, ctx),
	}
	err := service.Migrate()
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return &service
}

func (s *InventoryService) Migrate() error {
	if s.ctx.SkipMigration {
		return nil
	}
	if err := product.Migrate(s.ctx.DB); err != nil {
		fmt.Println("ERROR ACCOUNT", err)
		return err
	}
	if err := warehouse.Migrate(s.ctx.DB); err != nil {
		return err
	}
	if err := stockmovement.Migrate(s.ctx.DB); err != nil {
		return err
	}

	return nil
}
func (s *InventoryService) DB() *gorm.DB {
	return s.ctx.DB
}
