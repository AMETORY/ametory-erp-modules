package budget

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
)

type BudgetService struct {
	ctx *context.ERPContext
}

func NewBudgetService(ctx *context.ERPContext) *BudgetService {
	return &BudgetService{ctx: ctx}
}

func (s *BudgetService) CreateBudget(model *models.BudgetModel) (*models.BudgetModel, error) {
	return model, s.ctx.DB.Create(&model).Error
}

func (s *BudgetService) GetBudgetByID(id string) (*models.BudgetModel, error) {
	model := models.BudgetModel{}
	err := s.ctx.DB.
		Preload("BudgetKPIs").
		Preload("BudgetOutputs").
		Preload("BudgetComponents").
		Preload("BudgetActivities").
		Preload("BudgetStrategicObjectives").
		Where("id = ?", id).
		First(&model).
		Error
	return &model, err
}

func (s *BudgetService) UpdateBudget(model *models.BudgetModel) error {
	return s.ctx.DB.Model(&model).Where("id = ?", model.ID).Updates(model).Error

}

func (s *BudgetService) DeleteBudget(id string) error {
	return s.ctx.DB.Delete(&models.BudgetModel{}, "id = ?", id).Error
}
