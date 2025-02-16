package employee_activity

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"gorm.io/gorm"
)

type EmployeeActivityService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewEmployeeActivityService(ctx *context.ERPContext) *EmployeeActivityService {
	return &EmployeeActivityService{db: ctx.DB, ctx: ctx}
}
