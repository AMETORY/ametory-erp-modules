package payroll

import "github.com/AMETORY/ametory-erp-modules/shared/models"

func (s *PayrollService) CreatePayRollCost(cost *models.PayRollCostModel) error {
	return s.db.Create(cost).Error
}

func (s *PayrollService) GetPayRollCostByID(id string) (*models.PayRollCostModel, error) {
	var cost models.PayRollCostModel
	err := s.db.First(&cost, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &cost, nil
}

func (s *PayrollService) UpdatePayRollCost(cost *models.PayRollCostModel) error {
	return s.db.Save(cost).Error
}

func (s *PayrollService) DeletePayRollCost(id string) error {
	return s.db.Delete(&models.PayRollCostModel{}, "id = ?", id).Error
}
