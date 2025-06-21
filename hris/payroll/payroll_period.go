package payroll

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
)

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

func (s *PayrollService) FindAllPayRollPeriodes(request *http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.Model(&models.PayRollPeriodeModel{})
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.PayRollPeriodeModel{})
	page.Page = page.Page + 1
	return page, nil
}
