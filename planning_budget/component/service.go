package component

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/google/uuid"
)

type ComponentService struct {
	ctx *context.ERPContext
}

// NewComponentService creates a new instance of ComponentService.
//
// It takes an ERPContext as a parameter, which provides the necessary context for
// performing operations related to components, such as database interactions and authentication.
// Returns a pointer to the newly created ComponentService.

func NewComponentService(ctx *context.ERPContext) *ComponentService {
	return &ComponentService{ctx: ctx}
}

// CreateComponent creates a new component in the database.
//
// The method takes a pointer to a BudgetComponentModel as a parameter, which contains the necessary
// information to create a new component, such as name and description.
// It then uses the GORM library to create a new record in the budget_components table,
// and returns the error encountered during this process.
func (s *ComponentService) CreateComponent(component *models.BudgetComponentModel) error {
	return s.ctx.DB.Create(component).Error
}

// GetComponent retrieves a component from the database with the given ID.
//
// The method takes a uuid as a parameter, which is the unique identifier of the
// component to be retrieved. It uses the GORM library to query the database
// and retrieve the component with the given ID, and returns the component and
// any error encountered during this process.
func (s *ComponentService) GetComponent(id uuid.UUID) (*models.BudgetComponentModel, error) {
	component := &models.BudgetComponentModel{}
	err := s.ctx.DB.Where("id = ?", id).First(component).Error
	if err != nil {
		return nil, err
	}
	return component, nil
}

// UpdateComponent updates the details of an existing budget component.
//
// The method takes a string id and a pointer to a BudgetComponentModel as parameters.
// It uses the GORM library to update the component in the database with the given id.
// Returns an error if the update operation fails.
func (s *ComponentService) UpdateComponent(id string, component *models.BudgetComponentModel) error {
	return s.ctx.DB.Where("id = ?", id).Updates(component).Error
}

// DeleteComponent deletes a component from the database.
//
// The method takes the ID of the component to be deleted as a string.
// It uses the GORM library to delete the component from the database,
// and returns an error if the deletion operation fails.
func (s *ComponentService) DeleteComponent(id string) error {
	return s.ctx.DB.Where("id = ?", id).Delete(&models.BudgetComponentModel{}).Error
}
