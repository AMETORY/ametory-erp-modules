package branch

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"gorm.io/gorm"
)

type BranchService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewBranchService(db *gorm.DB, ctx *context.ERPContext) *BranchService {
	return &BranchService{db: db, ctx: ctx}
}
