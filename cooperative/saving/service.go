package saving

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"gorm.io/gorm"
)

type SavingService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewSavingService(db *gorm.DB, ctx *context.ERPContext) *SavingService {
	return &SavingService{
		db:  db,
		ctx: ctx,
	}
}
