package work_order

import "github.com/AMETORY/ametory-erp-modules/shared/models"

// CreateProductionOutput creates a new production output
//
// The function takes a production output as an argument and creates a new production output in the database.
//
// The function returns the created production output and an error if any.
func (s WorkOrderService) CreateProductionOutput(output *models.ProductionOutput) (*models.ProductionOutput, error) {
	if err := s.db.Create(output).Error; err != nil {
		return nil, err
	}

	return output, nil
}

// ReadProductionOutput reads a production output by id
//
// The function takes an id as an argument and reads the production output from the database.
//
// The function returns the read production output and an error if any.
func (s WorkOrderService) ReadProductionOutput(id string) (*models.ProductionOutput, error) {
	w := &models.ProductionOutput{}
	if err := s.db.Where("id = ?", id).First(w).Error; err != nil {
		return nil, err
	}

	return w, nil
}

// GetProductionOutputWithProductionProcess reads a production output with its production process by id
//
// The function takes an id as an argument and reads the production output with its production process from the database.
//
// The function returns the read production output and an error if any.
func (s WorkOrderService) GetProductionOutputWithProductionProcess(id string) (*models.ProductionOutput, error) {
	w := &models.ProductionOutput{}
	if err := s.db.Where("id = ?", id).Preload("ProductionProcess").First(w).Error; err != nil {
		return nil, err
	}

	return w, nil
}

// UpdateProductionOutput updates a production output
//
// The function takes a production output as an argument and updates the production output in the database.
//
// The function returns an error if any.
func (s WorkOrderService) UpdateProductionOutput(output *models.ProductionOutput) error {
	if err := s.db.Save(output).Error; err != nil {
		return err
	}

	return nil
}

// DeleteProductionOutput deletes a production output
//
// The function takes an id as an argument and deletes the production output from the database.
//
// The function returns an error if any.
func (s WorkOrderService) DeleteProductionOutput(id string) error {
	if err := s.db.Delete(&models.ProductionOutput{}, "id = ?", id).Error; err != nil {
		return err
	}

	return nil
}
