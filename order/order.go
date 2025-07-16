package order

import (
	"fmt"
	"log"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/finance"
	"github.com/AMETORY/ametory-erp-modules/inventory"
	"github.com/AMETORY/ametory-erp-modules/order/banner"
	"github.com/AMETORY/ametory-erp-modules/order/merchant"
	"github.com/AMETORY/ametory-erp-modules/order/payment"
	"github.com/AMETORY/ametory-erp-modules/order/payment_term"
	"github.com/AMETORY/ametory-erp-modules/order/pos"
	"github.com/AMETORY/ametory-erp-modules/order/promotion"
	"github.com/AMETORY/ametory-erp-modules/order/sales"
	"github.com/AMETORY/ametory-erp-modules/order/sales_return"
	"github.com/AMETORY/ametory-erp-modules/order/withdrawal"
	"gorm.io/gorm"
)

type OrderService struct {
	ctx                *context.ERPContext
	SalesService       *sales.SalesService
	PosService         *pos.POSService
	MerchantService    *merchant.MerchantService
	PaymentService     *payment.PaymentService
	WithdrawalService  *withdrawal.WithdrawalService
	BannerService      *banner.BannerService
	PromotionService   *promotion.PromotionService
	PaymentTermService *payment_term.PaymentTermService
	SalesReturnService *sales_return.SalesReturnService
}

// NewOrderService initializes a new OrderService instance.
//
// It takes an ERPContext and initializes sub-services. It also runs a migration
// for the database.
//
// If the migration fails, it will return nil.
func NewOrderService(ctx *context.ERPContext) *OrderService {
	fmt.Println("INIT ORDER SERVICE")
	var financeService *finance.FinanceService
	financeSrv, ok := ctx.FinanceService.(*finance.FinanceService)
	if ok {
		financeService = financeSrv
	}
	inventoryService := inventory.NewInventoryService(ctx)
	salesService := sales.NewSalesService(ctx.DB, ctx, financeService, inventoryService)
	var service = OrderService{
		ctx:                ctx,
		SalesService:       salesService,
		PosService:         pos.NewPOSService(ctx.DB, ctx, financeService),
		MerchantService:    merchant.NewMerchantService(ctx.DB, ctx, financeService, inventoryService),
		PaymentService:     payment.NewPaymentService(ctx.DB, ctx),
		WithdrawalService:  withdrawal.NewWithdrawalService(ctx.DB, ctx),
		BannerService:      banner.NewBannerService(ctx.DB, ctx),
		PromotionService:   promotion.NewPromotionService(ctx.DB, ctx),
		PaymentTermService: payment_term.NewPaymentTermService(ctx.DB, ctx),
		SalesReturnService: sales_return.NewSalesReturnService(ctx.DB, ctx, financeService, inventoryService.StockMovementService, salesService),
	}
	err := service.Migrate()
	if err != nil {
		fmt.Println("INIT ORDER SERVICE ERROR", err)
		return nil
	}
	return &service
}

// Migrate runs the database migrations for the order module.
//
// It first checks if the SkipMigration flag is set in the context. If it is, the
// function returns immediately.
//
// It then calls the Migrate functions of the Sales, POS, Merchant, Payment,
// Withdrawal, Banner, Promotion, and Payment Term services, passing the database
// connection from the context. If any of these calls returns an error, the
// function logs the error and returns it.
func (s *OrderService) Migrate() error {
	if s.ctx.SkipMigration {
		return nil
	}
	if err := sales.Migrate(s.ctx.DB); err != nil {
		log.Println("ERROR SALES", err)
		return err
	}
	if err := pos.Migrate(s.ctx.DB); err != nil {
		log.Println("ERROR POS", err)
		return err
	}
	if err := merchant.Migrate(s.ctx.DB); err != nil {
		log.Println("ERROR MERCHANT", err)
		return err
	}
	if err := payment.Migrate(s.ctx.DB); err != nil {
		log.Println("ERROR PAYMENT", err)
		return err
	}
	if err := withdrawal.Migrate(s.ctx.DB); err != nil {
		log.Println("ERROR WITHDRAWAL", err)
		return err
	}
	if err := banner.Migrate(s.ctx.DB); err != nil {
		log.Println("ERROR BANNER", err)
		return err
	}
	if err := promotion.Migrate(s.ctx.DB); err != nil {
		log.Println("ERROR PROMOTION", err)
		return err
	}
	if err := payment_term.Migrate(s.ctx.DB); err != nil {
		log.Println("ERROR PAYMENT TERM", err)
		return err
	}

	return nil
}

// DB returns the underlying database connection.
//
// The method returns the GORM database connection that is used by the service
// for CRUD (Create, Read, Update, Delete) operations.
func (s *OrderService) DB() *gorm.DB {
	return s.ctx.DB
}
