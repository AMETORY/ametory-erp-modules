package employee

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"gorm.io/gorm"
)

type EmployeeService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewEmployeeService(ctx *context.ERPContext) *EmployeeService {
	return &EmployeeService{db: ctx.DB, ctx: ctx}
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.EmployeeModel{},
	)
}
