package strategic_objective

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
)

type StrategicObjectiveService struct {
	ctx *context.ERPContext
}

func NewStrategicObjectiveService(ctx *context.ERPContext) *StrategicObjectiveService {
	return &StrategicObjectiveService{ctx: ctx}
}

func (s *StrategicObjectiveService) CreateStrategicObjective(objective *models.BudgetStrategicObjectiveModel) error {
	return s.ctx.DB.Create(objective).Error
}

func (s *StrategicObjectiveService) GetStrategicObjectiveByID(id string) (*models.BudgetStrategicObjectiveModel, error) {
	var objective models.BudgetStrategicObjectiveModel
	err := s.ctx.DB.First(&objective, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &objective, nil
}

func (s *StrategicObjectiveService) UpdateStrategicObjective(objective *models.BudgetStrategicObjectiveModel) error {
	return s.ctx.DB.Save(objective).Error
}

func (s *StrategicObjectiveService) DeleteStrategicObjective(id string) error {
	return s.ctx.DB.Delete(&models.BudgetStrategicObjectiveModel{}, "id = ?", id).Error
}
