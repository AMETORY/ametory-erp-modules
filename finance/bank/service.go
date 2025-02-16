package bank

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"gorm.io/gorm"
)

type BankService struct {
	ctx *context.ERPContext
	db  *gorm.DB
}

func NewBankService(db *gorm.DB, ctx *context.ERPContext) *BankService {
	return &BankService{ctx: ctx, db: db}
}
