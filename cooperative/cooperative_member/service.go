package cooperative_member

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"gorm.io/gorm"
)

type CooperativeMemberService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewCooperativeMemberService(db *gorm.DB, ctx *context.ERPContext) *CooperativeMemberService {
	return &CooperativeMemberService{
		db:  db,
		ctx: ctx,
	}
}
