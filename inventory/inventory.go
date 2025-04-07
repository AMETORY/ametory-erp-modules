package inventory

import (
	"fmt"
	"log"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/file"
	"github.com/AMETORY/ametory-erp-modules/finance"
	"github.com/AMETORY/ametory-erp-modules/inventory/brand"
	"github.com/AMETORY/ametory-erp-modules/inventory/product"
	"github.com/AMETORY/ametory-erp-modules/inventory/purchase"
	"github.com/AMETORY/ametory-erp-modules/inventory/purchase_return"
	stockmovement "github.com/AMETORY/ametory-erp-modules/inventory/stock_movement"
	"github.com/AMETORY/ametory-erp-modules/inventory/stock_opname"
	"github.com/AMETORY/ametory-erp-modules/inventory/unit"
	"github.com/AMETORY/ametory-erp-modules/inventory/warehouse"
	"gorm.io/gorm"
)

type InventoryService struct {
	ctx                     *context.ERPContext
	MasterProductService    *product.MasterProductService
	ProductService          *product.ProductService
	ProductCategoryService  *product.ProductCategoryService
	ProductAttributeService *product.ProductAttributeService
	PriceCategoryService    *product.PriceCategoryService
	WarehouseService        *warehouse.WarehouseService
	StockMovementService    *stockmovement.StockMovementService
	PurchaseService         *purchase.PurchaseService
	PurchaseReturnService   *purchase_return.PurchaseReturnService
	BrandService            *brand.BrandService
	StockOpenameService     *stock_opname.StockOpnameService
	TagService              *product.TagService
	UnitService             *unit.UnitService
}

func NewInventoryService(ctx *context.ERPContext) *InventoryService {
	fmt.Println("INIT INVENTORY SERVICE")
	var financeService *finance.FinanceService
	var fileService *file.FileService
	financeSrv, ok := ctx.FinanceService.(*finance.FinanceService)
	if ok {
		financeService = financeSrv
	}
	fileSrv, ok := ctx.FileService.(*file.FileService)
	if ok {
		fileService = fileSrv
	}
	stockmovementSrv := stockmovement.NewStockMovementService(ctx.DB, ctx)
	tagService := product.NewTagService(ctx.DB, ctx)
	unitService := unit.NewUnitService(ctx.DB, ctx)
	productSrv := product.NewProductService(ctx.DB, ctx, fileService, tagService)
	purchaseSrv := purchase.NewPurchaseService(ctx.DB, ctx, financeService, stockmovementSrv)

	var service = InventoryService{
		ctx:                     ctx,
		MasterProductService:    product.NewMasterProductService(ctx.DB, ctx),
		ProductService:          productSrv,
		ProductCategoryService:  product.NewProductCategoryService(ctx.DB, ctx),
		ProductAttributeService: product.NewProductAttributeService(ctx.DB, ctx),
		PriceCategoryService:    product.NewPriceCategoryService(ctx.DB, ctx),
		WarehouseService:        warehouse.NewWarehouseService(ctx.DB, ctx),
		StockMovementService:    stockmovementSrv,
		PurchaseService:         purchaseSrv,
		PurchaseReturnService:   purchase_return.NewPurchaseReturnService(ctx.DB, ctx, financeService, stockmovementSrv, purchaseSrv),
		BrandService:            brand.NewBrandService(ctx.DB, ctx),
		TagService:              tagService,
		StockOpenameService:     stock_opname.NewStockOpnameService(ctx.DB, ctx, productSrv, stockmovementSrv),
		UnitService:             unitService,
	}
	err := service.Migrate()
	if err != nil {
		panic(err)
	}
	return &service
}

func (s *InventoryService) Migrate() error {
	if s.ctx.SkipMigration {
		return nil
	}
	if err := product.Migrate(s.ctx.DB); err != nil {
		log.Println("ERROR MIGRATING PRODUCT", err)
		return err
	}
	if err := warehouse.Migrate(s.ctx.DB); err != nil {
		log.Println("ERROR MIGRATING WAREHOUSE", err)
		return err
	}
	if err := stockmovement.Migrate(s.ctx.DB); err != nil {
		return err
	}
	if err := brand.Migrate(s.ctx.DB); err != nil {
		log.Println("ERROR MIGRATING BRAND", err)
		return err
	}
	if err := purchase.Migrate(s.ctx.DB); err != nil {
		log.Println("ERROR MIGRATING PURCHASE", err)
		return err
	}
	if err := stock_opname.Migrate(s.ctx.DB); err != nil {
		log.Println("ERROR MIGRATING STOCK OPNAME", err)
		return err
	}
	if err := unit.Migrate(s.ctx.DB); err != nil {
		log.Println("ERROR MIGRATING UNIT", err)
		return err
	}
	if err := purchase_return.Migrate(s.ctx.DB); err != nil {
		log.Println("ERROR MIGRATING PURCHASE RETURN", err)
		return err
	}

	return nil
}
func (s *InventoryService) DB() *gorm.DB {
	return s.ctx.DB
}
