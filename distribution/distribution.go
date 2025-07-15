package distribution

import (
	"fmt"
	"log"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/distribution/cart"
	"github.com/AMETORY/ametory-erp-modules/distribution/distributor"
	"github.com/AMETORY/ametory-erp-modules/distribution/logistic"
	"github.com/AMETORY/ametory-erp-modules/distribution/offering"
	"github.com/AMETORY/ametory-erp-modules/distribution/order_request"
	"github.com/AMETORY/ametory-erp-modules/distribution/shipping"
	"github.com/AMETORY/ametory-erp-modules/distribution/storage"
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
	LogisticService     *logistic.LogisticService
	StorageService      *storage.StorageService
}

// NewDistributionService creates a new instance of DistributionService with the given ERP context, audit trail service,
// inventory service, and order service.
//
// It also migrates the database according to the latest schema, and creates all the necessary subservices.
func NewDistributionService(ctx *context.ERPContext, auditTrailService *audit_trail.AuditTrailService, inventoryService *inventory.InventoryService, orderService *order.OrderService) *DistributionService {
	fmt.Println("INIT DISTRIBUTION SERVICE")

	var service = DistributionService{
		ctx:               ctx,
		auditTrailService: auditTrailService,
	}
	err := service.Migrate()
	if err != nil {
		panic(err)
	}
	service.LogisticService = logistic.NewLogisticService(ctx.DB, ctx, inventoryService)
	service.StorageService = storage.NewStorageService(ctx.DB, ctx, inventoryService)
	service.DistributorService = distributor.NewDistributorService(ctx.DB, ctx)
	service.OrderRequestService = order_request.NewOrderRequestService(ctx.DB, ctx, orderService.MerchantService, inventoryService.ProductService, auditTrailService)
	service.OfferingService = offering.NewOfferingService(ctx.DB, ctx, auditTrailService)
	service.ShippingService = shipping.NewShippingService(ctx.DB, ctx)
	service.CartService = cart.NewCartService(ctx.DB, ctx, inventoryService)

	return &service
}

// Migrate performs the database migration for the DistributionService.
//
// This method migrates the database tables for various distribution-related
// services, including logistic, storage, distributor, order request, offering,
// shipping, and cart. If the SkipMigration flag is set to true in the context,
// the migration process is skipped. It returns an error if any of the migration
// tasks fail, logging the error for each failed migration.

func (s *DistributionService) Migrate() error {
	if s.ctx.SkipMigration {
		return nil
	}
	if err := logistic.Migrate(s.ctx.DB); err != nil {
		log.Println("ERROR LOGISTIC MIGRATE", err)
		return err
	}
	if err := storage.Migrate(s.ctx.DB); err != nil {
		log.Println("ERROR STORAGE MIGRATE", err)
		return err
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
