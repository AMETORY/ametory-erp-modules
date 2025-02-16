package attendance

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"gorm.io/gorm"
)

type AttendanceService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewAttendanceService(ctx *context.ERPContext) *AttendanceService {
	return &AttendanceService{db: ctx.DB, ctx: ctx}
}
