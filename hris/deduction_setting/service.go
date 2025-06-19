package deduction_setting

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"gorm.io/gorm"
)

type DeductionSettingService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewDeductionSettingService(ctx *context.ERPContext) *DeductionSettingService {
	return &DeductionSettingService{db: ctx.DB, ctx: ctx}
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.DeductionSettingModel{},
	)
}
