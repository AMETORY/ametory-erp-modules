package bom

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"gorm.io/gorm"
)

type BuildOfMaterialService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewBuildOfMaterialService(db *gorm.DB, ctx *context.ERPContext) *BuildOfMaterialService {
	return &BuildOfMaterialService{
		db:  db,
		ctx: ctx,
	}
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&models.BillOfMaterial{}, &models.BOMItem{}, &models.BOMOperation{})
}
