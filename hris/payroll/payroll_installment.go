package payroll

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
)

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

func (s *PayrollService) FindAll(request *http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.Model(&models.PayRollInstallment{})
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.PayRollInstallment{})
	page.Page = page.Page + 1
	return page, nil
}
