package attendance_policy

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"gorm.io/gorm"
)

type AttendancePolicyService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewAttendancePolicyService(ctx *context.ERPContext) *AttendancePolicyService {
	return &AttendancePolicyService{db: ctx.DB, ctx: ctx}
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.AttendancePolicy{},
	)
}
