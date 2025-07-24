package output

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
)

type OutputService struct {
	ctx *context.ERPContext
}

func NewOutputService(ctx *context.ERPContext) *OutputService {
	return &OutputService{ctx: ctx}
}

func (s *OutputService) CreateBudgetOutput(kpi *models.BudgetOutputModel) error {
	return s.ctx.DB.Create(kpi).Error
}

func (s *OutputService) FindAllBudgetOutputs() ([]models.BudgetOutputModel, error) {
	var kpis []models.BudgetOutputModel
	err := s.ctx.DB.Find(&kpis).Error
	return kpis, err
}

func (s *OutputService) FindBudgetOutputByID(id string) (*models.BudgetOutputModel, error) {
	var kpi models.BudgetOutputModel
	err := s.ctx.DB.Where("id = ?", id).First(&kpi).Error
	return &kpi, err
}

func (s *OutputService) UpdateBudgetOutput(kpi *models.BudgetOutputModel) error {
	return s.ctx.DB.Save(kpi).Error
}

func (s *OutputService) DeleteBudgetOutput(kpi *models.BudgetOutputModel) error {
	return s.ctx.DB.Delete(kpi).Error
}
