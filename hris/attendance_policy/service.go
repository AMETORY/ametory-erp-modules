package attendance_policy

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"gorm.io/gorm"
)

type AttendancePolicyService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewAttendancePolicyService(ctx *context.ERPContext) *AttendancePolicyService {
	return &AttendancePolicyService{db: ctx.DB, ctx: ctx}
}
