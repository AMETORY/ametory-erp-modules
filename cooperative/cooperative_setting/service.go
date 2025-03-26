package cooperative_setting

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"gorm.io/gorm"
)

type CooperativeSettingService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewCooperativeSettingService(db *gorm.DB, ctx *context.ERPContext) *CooperativeSettingService {
	return &CooperativeSettingService{
		db:  db,
		ctx: ctx,
	}
}
