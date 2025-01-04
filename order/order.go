package order

import (
	"fmt"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/order/sales"
	"gorm.io/gorm"
)

type OrderService struct {
	ctx          *context.ERPContext
	SalesService *sales.SalesService
}

func NewOrderService(ctx *context.ERPContext) *OrderService {
	fmt.Println("INIT ORDER SERVICE")
	var service = OrderService{
		ctx:          ctx,
		SalesService: sales.NewSalesService(ctx.DB, ctx),
	}
	err := service.Migrate()
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return &service
}

func (s *OrderService) Migrate() error {
	if s.ctx.SkipMigration {
		return nil
	}
	if err := sales.Migrate(s.ctx.DB); err != nil {
		fmt.Println("ERROR ACCOUNT", err)
		return err
	}

	return nil
}

func (s *OrderService) DB() *gorm.DB {
	return s.ctx.DB
}
