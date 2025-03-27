package report

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/finance/report/report_object"
	"gorm.io/gorm"
)

type FinanceReportService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewFinanceReportService(db *gorm.DB, ctx *context.ERPContext) *FinanceReportService {
	return &FinanceReportService{
		db:  db,
		ctx: ctx,
	}
}

func (s *FinanceReportService) GenerateProfitLoss(report *report_object.ProfitLoss) error {

	return nil
}
