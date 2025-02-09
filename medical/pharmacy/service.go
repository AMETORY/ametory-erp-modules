package pharmacy

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"gorm.io/gorm"
)

// PharmacyService is the service for Pharmacy model
type PharmacyService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

// NewPharmacyService will create new PharmacyService
func NewPharmacyService(db *gorm.DB, ctx *context.ERPContext) *PharmacyService {
	return &PharmacyService{db: db, ctx: ctx}
}
