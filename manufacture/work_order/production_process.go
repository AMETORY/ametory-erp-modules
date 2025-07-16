package work_order

import "github.com/AMETORY/ametory-erp-modules/shared/models"

// CreateProductionProcess creates a new production process
//
// The function takes a production process as an argument and creates a new production process in the database.
//
// The function returns the created production process and an error if any.
func (s WorkOrderService) CreateProductionProcess(process *models.ProductionProcess) (*models.ProductionProcess, error) {
	if err := s.db.Create(process).Error; err != nil {
		return nil, err
	}

	return process, nil
}

// GetProductionProcess reads a production process by id
//
// The function takes an id as an argument and reads the production process from the database.
//
// The function returns the read production process and an error if any.
func (s WorkOrderService) GetProductionProcess(id string) (*models.ProductionProcess, error) {
	result := &models.ProductionProcess{}
	if err := s.db.Where("id = ?", id).Preload("AdditionalCosts").First(result).Error; err != nil {
		return nil, err
	}

	return result, nil
}

// UpdateProductionProcess updates a production process
//
// The function takes a production process as an argument and updates the production process in the database.
//
// The function returns an error if any.
func (s WorkOrderService) UpdateProductionProcess(process *models.ProductionProcess) error {
	if err := s.db.Save(process).Error; err != nil {
		return err
	}

	return nil
}

// DeleteProductionProcess deletes a production process
//
// The function takes an id as an argument and deletes the production process from the database.
//
// The function returns an error if any.
func (s WorkOrderService) DeleteProductionProcess(id string) error {
	if err := s.db.Delete(&models.ProductionProcess{}, "id = ?", id).Error; err != nil {
		return err
	}

	return nil
}

// CalculateTotalCost calculates the total cost of a production process
//
// The function takes a production process id as an argument, retrieves the production process with its components,
// and calculates the total cost based on the components' costs.
//
// The function returns the total cost and an error if any.
func (s WorkOrderService) CalculateProcessTotalCost(id string) (float64, error) {
	process, err := s.GetProductionProcess(id)
	if err != nil {
		return 0, err
	}

	totalCost := 0.0
	for _, component := range process.AdditionalCosts {
		totalCost += component.Amount
	}

	return totalCost, nil
}

// AddCost adds a new cost to a production process
//
// The function takes a production process id and a cost as arguments and adds a new cost to the production process in the database.
//
// The function returns an error if any.
func (s WorkOrderService) AddCost(id string, cost *models.ProductionAdditionalCost) error {
	process, err := s.GetProductionProcess(id)
	if err != nil {
		return err
	}

	process.AdditionalCosts = append(process.AdditionalCosts, *cost)
	if err := s.db.Save(process).Error; err != nil {
		return err
	}
	totalCost, err := s.CalculateProcessTotalCost(id)
	if err != nil {
		return err
	}

	process.TotalCost = totalCost

	if err := s.db.Save(process).Error; err != nil {
		return err
	}

	return nil
}

// EditCost edits a cost in a production process
//
// The function takes a production process id and a cost as arguments and edits the cost in the production process in the database.
//
// The function returns an error if any.
func (s WorkOrderService) EditCost(id string, cost *models.ProductionAdditionalCost) error {
	process, err := s.GetProductionProcess(id)
	if err != nil {
		return err
	}

	for i, c := range process.AdditionalCosts {
		if c.ID == cost.ID {
			process.AdditionalCosts[i] = *cost
			break
		}
	}

	if err := s.db.Save(process).Error; err != nil {
		return err
	}
	totalCost, err := s.CalculateProcessTotalCost(id)
	if err != nil {
		return err
	}

	process.TotalCost = totalCost

	if err := s.db.Save(process).Error; err != nil {
		return err
	}

	return nil
}

// DeleteCost deletes a cost from a production process
//
// The function takes a production process id and a cost id as arguments and deletes the cost from the production process in the database.
//
// The function returns an error if any.
func (s WorkOrderService) DeleteCost(id string, costID string) error {
	process, err := s.GetProductionProcess(id)
	if err != nil {
		return err
	}

	for i, c := range process.AdditionalCosts {
		if c.ID == costID {
			process.AdditionalCosts = append(process.AdditionalCosts[:i], process.AdditionalCosts[i+1:]...)
			break
		}
	}

	if err := s.db.Save(process).Error; err != nil {
		return err
	}

	totalCost, err := s.CalculateProcessTotalCost(id)
	if err != nil {
		return err
	}

	process.TotalCost = totalCost

	if err := s.db.Save(process).Error; err != nil {
		return err
	}

	return nil
}
