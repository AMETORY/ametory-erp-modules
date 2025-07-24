package kpi

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
)

type KPIService struct {
	ctx *context.ERPContext
}

// NewKPIService creates a new instance of KPIService.
//
// It takes an ERPContext as a parameter, which provides the necessary context for
// performing operations related to KPIs, such as database interactions and authentication.
// Returns a pointer to the newly created KPIService.

func NewKPIService(ctx *context.ERPContext) *KPIService {
	return &KPIService{ctx: ctx}
}

// CreateBudgetKPI creates a new BudgetKPIModel and saves it to the database.
//
// It takes a pointer to a BudgetKPIModel as a parameter, which is the KPI to be
// created. The function will return an error if the creation was unsuccessful.
func (s *KPIService) CreateBudgetKPI(kpi *models.BudgetKPIModel) error {
	return s.ctx.DB.Create(kpi).Error
}

// FindAllBudgetKPIs retrieves all BudgetKPIModel objects from the database.
//
// This function will return a slice of BudgetKPIModel objects, and an error if the
// retrieval was unsuccessful.
func (s *KPIService) FindAllBudgetKPIs() ([]models.BudgetKPIModel, error) {
	var kpis []models.BudgetKPIModel
	err := s.ctx.DB.Find(&kpis).Error
	return kpis, err
}

// FindBudgetKPIByID retrieves a BudgetKPIModel by its ID from the database.
//
// It takes an ID string as a parameter and returns a pointer to the BudgetKPIModel
// and an error if the retrieval was unsuccessful.
func (s *KPIService) FindBudgetKPIByID(id string) (*models.BudgetKPIModel, error) {
	var kpi models.BudgetKPIModel
	err := s.ctx.DB.Where("id = ?", id).First(&kpi).Error
	return &kpi, err
}

// UpdateBudgetKPI updates an existing BudgetKPIModel in the database.
//
// It takes a pointer to a BudgetKPIModel as a parameter, which contains the updated
// KPI data. The function returns an error if the update operation fails.
func (s *KPIService) UpdateBudgetKPI(kpi *models.BudgetKPIModel) error {
	return s.ctx.DB.Save(kpi).Error
}

// DeleteBudgetKPI deletes an existing BudgetKPIModel from the database.
//
// It takes a pointer to a BudgetKPIModel as a parameter, representing the KPI to be deleted.
// The function returns an error if the delete operation fails.
func (s *KPIService) DeleteBudgetKPI(kpi *models.BudgetKPIModel) error {
	return s.ctx.DB.Delete(kpi).Error
}
