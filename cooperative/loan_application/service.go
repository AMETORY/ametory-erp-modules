package loan_application

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"gorm.io/gorm"
)

type LoanApplicationService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewLoanApplicationService(db *gorm.DB, ctx *context.ERPContext) *LoanApplicationService {
	return &LoanApplicationService{
		db:  db,
		ctx: ctx,
	}
}
