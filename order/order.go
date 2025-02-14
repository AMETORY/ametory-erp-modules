package order

import (
	"fmt"
	"log"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/finance"
	"github.com/AMETORY/ametory-erp-modules/inventory"
	"github.com/AMETORY/ametory-erp-modules/order/merchant"
	"github.com/AMETORY/ametory-erp-modules/order/payment"
	"github.com/AMETORY/ametory-erp-modules/order/pos"
	"github.com/AMETORY/ametory-erp-modules/order/sales"
	"github.com/AMETORY/ametory-erp-modules/order/withdrawal"
	"gorm.io/gorm"
)

type OrderService struct {
	ctx               *context.ERPContext
	SalesService      *sales.SalesService
	PosService        *pos.POSService
	MerchantService   *merchant.MerchantService
	PaymentService    *payment.PaymentService
	WithdrawalService *withdrawal.WithdrawalService
}

func NewOrderService(ctx *context.ERPContext) *OrderService {
	fmt.Println("INIT ORDER SERVICE")
	var financeService *finance.FinanceService
	financeSrv, ok := ctx.FinanceService.(*finance.FinanceService)
	if ok {
		financeService = financeSrv
	}
	inventoryService := inventory.NewInventoryService(ctx)

	var service = OrderService{
		ctx:               ctx,
		SalesService:      sales.NewSalesService(ctx.DB, ctx, financeService),
		PosService:        pos.NewPOSService(ctx.DB, ctx, financeService),
		MerchantService:   merchant.NewMerchantService(ctx.DB, ctx, financeService, inventoryService),
		PaymentService:    payment.NewPaymentService(ctx.DB, ctx),
		WithdrawalService: withdrawal.NewWithdrawalService(ctx.DB, ctx),
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

	return nil
}

func (s *OrderService) DB() *gorm.DB {
	return s.ctx.DB
}
