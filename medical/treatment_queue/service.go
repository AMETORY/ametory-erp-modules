package treatment_queue

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"gorm.io/gorm"
)

// TreatmentQueueService is the service for treatment_queue model
type TreatmentQueueService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

// NewTreatmentQueueService creates new TreatmentQueueService
func NewTreatmentQueueService(db *gorm.DB, ctx *context.ERPContext) *TreatmentQueueService {
	return &TreatmentQueueService{db: db, ctx: ctx}
}
