package patient

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"gorm.io/gorm"
)

type PatientService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewPatientService(db *gorm.DB, ctx *context.ERPContext) *PatientService {
	return &PatientService{db: db, ctx: ctx}
}
