package employee_overtime

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"gorm.io/gorm"
)

type EmployeeOvertimeService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewEmployeeOvertimeService(ctx *context.ERPContext) *EmployeeOvertimeService {
	return &EmployeeOvertimeService{db: ctx.DB, ctx: ctx}
}
