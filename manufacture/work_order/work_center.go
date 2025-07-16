package work_order

import "github.com/AMETORY/ametory-erp-modules/shared/models"

// CreateWorkCenter creates a new work center
//
// The function takes a work center as an argument and creates a new work center in the database.
//
// The function returns the created work center and an error if any.
func (s WorkOrderService) CreateWorkCenter(w *models.WorkCenter) (*models.WorkCenter, error) {
	if err := s.db.Create(w).Error; err != nil {
		return nil, err
	}

	return w, nil
}

// GetWorkCenterByID retrieves a work center from the database by ID.
//
// It takes an ID as input and returns a pointer to a WorkCenterModel and an error.
// The function uses GORM to retrieve the work center data from the work_centers table.
// If the operation fails, an error is returned.
func (s WorkOrderService) GetWorkCenterByID(id string) (*models.WorkCenter, error) {
	w := &models.WorkCenter{}
	if err := s.db.Where("id = ?", id).First(w).Error; err != nil {
		return nil, err
	}

	return w, nil
}

// UpdateWorkCenter updates a work center
//
// The function takes a work center as an argument and updates the work center in the database.
//
// The function returns an error if any.
func (s WorkOrderService) UpdateWorkCenter(w *models.WorkCenter) error {
	if err := s.db.Save(w).Error; err != nil {
		return err
	}

	return nil
}

// DeleteWorkCenter deletes a work center in the database.
//
// The function takes a work center ID and attempts to delete the work center
// with the given ID from the database. If the deletion is successful, the
// function returns nil. If the deletion operation fails, the function returns
// an error.
func (s WorkOrderService) DeleteWorkCenter(id string) error {
	if err := s.db.Delete(&models.WorkCenter{}, "id = ?", id).Error; err != nil {
		return err
	}

	return nil
}
