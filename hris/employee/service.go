package employee

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"gorm.io/gorm"
)

type EmployeeService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewEmployeeService(ctx *context.ERPContext) *EmployeeService {
	return &EmployeeService{db: ctx.DB, ctx: ctx}
}
