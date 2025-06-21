package payroll

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
)

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

func (s *PayrollService) FindAllPayRollCosts(request *http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.Model(&models.PayRollCostModel{})
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.PayRollCostModel{})
	page.Page = page.Page + 1
	return page, nil
}
