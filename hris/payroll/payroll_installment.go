package payroll

import "github.com/AMETORY/ametory-erp-modules/shared/models"

func (s *PayrollService) CreatePayRollInstallment(installment *models.PayRollInstallment) error {
	return s.db.Create(installment).Error
}

func (s *PayrollService) GetPayRollInstallmentByID(id string) (*models.PayRollInstallment, error) {
	var installment models.PayRollInstallment
	err := s.db.First(&installment, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &installment, nil
}

func (s *PayrollService) UpdatePayRollInstallment(installment *models.PayRollInstallment) error {
	return s.db.Save(installment).Error
}

func (s *PayrollService) DeletePayRollInstallment(id string) error {
	return s.db.Delete(&models.PayRollInstallment{}, "id = ?", id).Error
}
