package output

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
)

type OutputService struct {
	ctx *context.ERPContext
}

// NewOutputService creates a new instance of OutputService.
//
// It takes an ERPContext as parameter and returns a pointer to a OutputService.
func NewOutputService(ctx *context.ERPContext) *OutputService {
	return &OutputService{ctx: ctx}
}

// CreateBudgetOutput creates a new budget output in the database.
//
// It takes a pointer to a BudgetOutputModel as parameter and returns an error
// if the creation was unsuccessful.
func (s *OutputService) CreateBudgetOutput(kpi *models.BudgetOutputModel) error {
	return s.ctx.DB.Create(kpi).Error
}

// FindAllBudgetOutputs retrieves all BudgetOutputModel objects from the database.
//
// This function returns a slice of BudgetOutputModel objects and an error if the
// retrieval was unsuccessful.
func (s *OutputService) FindAllBudgetOutputs() ([]models.BudgetOutputModel, error) {
	var kpis []models.BudgetOutputModel
	err := s.ctx.DB.Find(&kpis).Error
	return kpis, err
}

// FindBudgetOutputByID retrieves a BudgetOutputModel by its ID from the database.
//
// It takes an ID string as a parameter and returns a pointer to the BudgetOutputModel
// and an error if the retrieval was unsuccessful.
func (s *OutputService) FindBudgetOutputByID(id string) (*models.BudgetOutputModel, error) {
	var kpi models.BudgetOutputModel
	err := s.ctx.DB.Where("id = ?", id).First(&kpi).Error
	return &kpi, err
}

// UpdateBudgetOutput updates an existing budget output in the database.
//
// It takes a pointer to a BudgetOutputModel as parameter and returns an error
// if the update operation fails.
func (s *OutputService) UpdateBudgetOutput(kpi *models.BudgetOutputModel) error {
	return s.ctx.DB.Save(kpi).Error
}

// DeleteBudgetOutput deletes an existing budget output from the database.
//
// It takes a pointer to a BudgetOutputModel as parameter, representing the budget output to be deleted.
// The function returns an error if the delete operation fails.
func (s *OutputService) DeleteBudgetOutput(kpi *models.BudgetOutputModel) error {
	return s.ctx.DB.Delete(kpi).Error
}
