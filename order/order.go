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

func (s *OrderService) DB() *gorm.DB {
	return s.ctx.DB
}
