package payroll

import "github.com/AMETORY/ametory-erp-modules/shared/models"

func (s *PayrollService) CreatePayRollPeriode(periode *models.PayRollPeriodeModel) error {
	return s.db.Create(periode).Error
}

func (s *PayrollService) GetPayRollPeriodeByID(id string) (*models.PayRollPeriodeModel, error) {
	var periode models.PayRollPeriodeModel
	err := s.db.Preload("PayRolls").First(&periode, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &periode, nil
}

func (s *PayrollService) UpdatePayRollPeriode(periode *models.PayRollPeriodeModel) error {
	return s.db.Save(periode).Error
}

func (s *PayrollService) DeletePayRollPeriode(id string) error {
	return s.db.Delete(&models.PayRollPeriodeModel{}, "id = ?", id).Error
}
