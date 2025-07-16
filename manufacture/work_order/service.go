package work_order

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"gorm.io/gorm"
)

type WorkOrderService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewWorkOrderService(db *gorm.DB, ctx *context.ERPContext) *WorkOrderService {
	return &WorkOrderService{
		db:  db,
		ctx: ctx,
	}
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&models.WorkOrder{}, &models.ProductionProcess{}, &models.ProductionAdditionalCost{}, &models.ProductionOutput{}, &models.WorkCenter{})
}

// CreateWorkOrder creates a new work order
//
// The function takes a work order and its components as an argument and creates a new work order in the database.
//
// The function returns the created work order and an error if any.
func (s WorkOrderService) CreateWorkOrder(w *models.WorkOrder) (*models.WorkOrder, error) {
	if err := s.db.Create(w).Error; err != nil {
		return nil, err
	}

	return w, nil
}

// ReadWorkOrder reads a work order by id
//
// The function takes an id as an argument and reads the work order from the database.
//
// The function returns the read work order and an error if any.
func (s WorkOrderService) ReadWorkOrder(id string) (*models.WorkOrder, error) {
	w := &models.WorkOrder{}
	if err := s.db.Where("id = ?", id).First(w).Error; err != nil {
		return nil, err
	}

	return w, nil
}

// GetWorkOrderWithProductionProcess reads a work order with its components by id
//
// The function takes an id as an argument and reads the work order with its components from the database.
//
// The function returns the read work order and an error if any.
func (s WorkOrderService) GetWorkOrderWithProductionProcess(id string) (*models.WorkOrder, error) {
	w := &models.WorkOrder{}
	if err := s.db.Where("id = ?", id).Preload("ProductionProcess").First(w).Error; err != nil {
		return nil, err
	}

	return w, nil
}

// UpdateWorkOrder updates a work order
//
// The function takes a work order and its components as an argument and updates the work order in the database.
//
// The function returns an error if any.
func (s WorkOrderService) UpdateWorkOrder(w *models.WorkOrder) error {
	if err := s.db.Save(w).Error; err != nil {
		return err
	}

	return nil
}

// DeleteWorkOrder deletes a work order
//
// The function takes an id as an argument and deletes the work order from the database.
//
// The function returns an error if any.
func (s WorkOrderService) DeleteWorkOrder(id string) error {
	if err := s.db.Delete(&models.WorkOrder{}, "id = ?", id).Error; err != nil {
		return err
	}

	return nil
}

// CalculateTotalCost calculates the total cost of a work order
//
// The function takes a work order ID as an argument, retrieves the work order with its components,
// and calculates the total cost based on the components' costs.
//
// The function returns the total cost and an error if any.
func (s WorkOrderService) CalculateTotalCost(id string) (float64, error) {
	w, err := s.GetWorkOrderWithProductionProcess(id)
	if err != nil {
		return 0, err
	}

	totalCost := 0.0
	for _, component := range w.ProductionProcesses {
		totalCost += component.TotalCost
	}

	return totalCost, nil
}
