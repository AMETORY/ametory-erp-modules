package strategic_objective

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
)

type StrategicObjectiveService struct {
	ctx *context.ERPContext
}

// NewStrategicObjectiveService creates a new instance of StrategicObjectiveService.
//
// It takes an ERPContext as an argument and returns a pointer to a StrategicObjectiveService.
// This service is responsible for performing operations related to strategic objectives
// within the planning budget module.

func NewStrategicObjectiveService(ctx *context.ERPContext) *StrategicObjectiveService {
	return &StrategicObjectiveService{ctx: ctx}
}

// CreateStrategicObjective creates a new strategic objective in the database.
//
// It takes a pointer to a BudgetStrategicObjectiveModel as an argument and
// returns an error if the creation fails.
func (s *StrategicObjectiveService) CreateStrategicObjective(objective *models.BudgetStrategicObjectiveModel) error {
	return s.ctx.DB.Create(objective).Error
}

// GetStrategicObjectiveByID retrieves a strategic objective by its ID.
//
// It takes a string ID as argument and returns a pointer to a
// BudgetStrategicObjectiveModel and an error if the query fails.
func (s *StrategicObjectiveService) GetStrategicObjectiveByID(id string) (*models.BudgetStrategicObjectiveModel, error) {
	var objective models.BudgetStrategicObjectiveModel
	err := s.ctx.DB.First(&objective, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &objective, nil
}

// UpdateStrategicObjective updates an existing strategic objective in the database.
//
// It takes a pointer to a BudgetStrategicObjectiveModel as input and attempts
// to update the corresponding record in the database. If the update fails,
// it returns an error. Otherwise, it returns nil on successful update.

func (s *StrategicObjectiveService) UpdateStrategicObjective(objective *models.BudgetStrategicObjectiveModel) error {
	return s.ctx.DB.Save(objective).Error
}

// DeleteStrategicObjective deletes a strategic objective from the database.
//
// It takes a string ID as an argument, which specifies the strategic objective
// to be deleted. The function returns an error if the deletion fails; otherwise,
// it returns nil upon successful deletion.

func (s *StrategicObjectiveService) DeleteStrategicObjective(id string) error {
	return s.ctx.DB.Delete(&models.BudgetStrategicObjectiveModel{}, "id = ?", id).Error
}
