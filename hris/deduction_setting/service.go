package deduction_setting

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"gorm.io/gorm"
)

type DeductionSettingService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewDeductionSettingService(ctx *context.ERPContext) *DeductionSettingService {
	return &DeductionSettingService{db: ctx.DB, ctx: ctx}
}
