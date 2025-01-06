package distribution

import (
	"fmt"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/distribution/distributor"
	"gorm.io/gorm"
)

type DistributionService struct {
	ctx                *context.ERPContext
	DistributorService *distributor.DistributorService
}

func NewDistributionService(ctx *context.ERPContext) *DistributionService {
	fmt.Println("INIT DISTRIBUTION SERVICE")

	var service = DistributionService{
		ctx: ctx,
	}
	service.DistributorService = distributor.NewDistributorService(ctx.DB, ctx)
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
		return err
	}

	return nil
}

func (s *DistributionService) DB() *gorm.DB {
	return s.ctx.DB
}
