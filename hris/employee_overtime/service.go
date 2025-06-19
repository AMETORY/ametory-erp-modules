package employee_overtime

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"gorm.io/gorm"
)

type EmployeeOvertimeService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewEmployeeOvertimeService(ctx *context.ERPContext) *EmployeeOvertimeService {
	return &EmployeeOvertimeService{db: ctx.DB, ctx: ctx}
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.EmployeeOvertimeModel{},
	)
}
