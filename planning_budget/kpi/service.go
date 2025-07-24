package kpi

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
)

type KPIService struct {
	ctx *context.ERPContext
}

func NewKPIService(ctx *context.ERPContext) *KPIService {
	return &KPIService{ctx: ctx}
}
func (s *KPIService) CreateBudgetKPI(kpi *models.BudgetKPIModel) error {
	return s.ctx.DB.Create(kpi).Error
}

func (s *KPIService) FindAllBudgetKPIs() ([]models.BudgetKPIModel, error) {
	var kpis []models.BudgetKPIModel
	err := s.ctx.DB.Find(&kpis).Error
	return kpis, err
}

func (s *KPIService) FindBudgetKPIByID(id string) (*models.BudgetKPIModel, error) {
	var kpi models.BudgetKPIModel
	err := s.ctx.DB.Where("id = ?", id).First(&kpi).Error
	return &kpi, err
}

func (s *KPIService) UpdateBudgetKPI(kpi *models.BudgetKPIModel) error {
	return s.ctx.DB.Save(kpi).Error
}

func (s *KPIService) DeleteBudgetKPI(kpi *models.BudgetKPIModel) error {
	return s.ctx.DB.Delete(kpi).Error
}
