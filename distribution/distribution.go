package distribution

import (
	"fmt"
	"log"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/distribution/cart"
	"github.com/AMETORY/ametory-erp-modules/distribution/distributor"
	"github.com/AMETORY/ametory-erp-modules/distribution/offering"
	"github.com/AMETORY/ametory-erp-modules/distribution/order_request"
	"github.com/AMETORY/ametory-erp-modules/distribution/shipping"
	"github.com/AMETORY/ametory-erp-modules/inventory"
	"github.com/AMETORY/ametory-erp-modules/order"
	"github.com/AMETORY/ametory-erp-modules/shared/audit_trail"
	"gorm.io/gorm"
)

type DistributionService struct {
	ctx                 *context.ERPContext
	DistributorService  *distributor.DistributorService
	OfferingService     *offering.OfferingService
	OrderRequestService *order_request.OrderRequestService
	ShippingService     *shipping.ShippingService
	auditTrailService   *audit_trail.AuditTrailService
	CartService         *cart.CartService
}

func NewDistributionService(ctx *context.ERPContext, auditTrailService *audit_trail.AuditTrailService, inventoryService *inventory.InventoryService, orderService *order.OrderService) *DistributionService {
	fmt.Println("INIT DISTRIBUTION SERVICE")

	var service = DistributionService{
		ctx:               ctx,
		auditTrailService: auditTrailService,
	}
	service.DistributorService = distributor.NewDistributorService(ctx.DB, ctx)
	service.OrderRequestService = order_request.NewOrderRequestService(ctx.DB, ctx, orderService.MerchantService, inventoryService.ProductService, auditTrailService)
	service.OfferingService = offering.NewOfferingService(ctx.DB, ctx, auditTrailService)
	service.ShippingService = shipping.NewShippingService(ctx.DB, ctx)
	service.CartService = cart.NewCartService(ctx.DB, ctx, inventoryService)
	err := service.Migrate()
	if err != nil {
		panic(err)
	}
	return &service
}

func (s *DistributionService) Migrate() error {
	if s.ctx.SkipMigration {
		return nil
	}
	if err := distributor.Migrate(s.ctx.DB); err != nil {
		log.Println("ERROR DISTRIBUTOR MIGRATE", err)
		return err
	}
	if err := order_request.Migrate(s.ctx.DB); err != nil {
		log.Println("ERROR ORDER REQUEST MIGRATE", err)
		return err
	}
	if err := offering.Migrate(s.ctx.DB); err != nil {
		log.Println("ERROR OFFERING MIGRATE", err)
		return err
	}

	if err := shipping.Migrate(s.ctx.DB); err != nil {
		log.Println("ERROR SHIPPING MIGRATE", err)
		return err
	}
	if err := cart.Migrate(s.ctx.DB); err != nil {
		log.Println("ERROR CART MIGRATE", err)
		return err
	}

	return nil
}

func (s *DistributionService) DB() *gorm.DB {
	return s.ctx.DB
}
