package work_order

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"gorm.io/gorm"
)

type WorkOrderService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewWorkOrderService(db *gorm.DB, ctx *context.ERPContext) *WorkOrderService {
	return &WorkOrderService{
		db:  db,
		ctx: ctx,
	}
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&models.WorkOrder{}, &models.ProductionProcess{}, &models.ProductionAdditionalCost{}, &models.ProductionOutput{}, &models.WorkCenter{})
}
