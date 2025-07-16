package payroll

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
)

// CreatePayRollInstallment creates a new installment for a payroll.
// The installment must not exist in the database before calling this method.
func (s *PayrollService) CreatePayRollInstallment(installment *models.PayRollInstallment) error {
	return s.db.Create(installment).Error
}

// GetPayRollInstallmentByID returns the installment with given id
// or an error if none is found
func (s *PayrollService) GetPayRollInstallmentByID(id string) (*models.PayRollInstallment, error) {
	var installment models.PayRollInstallment
	err := s.db.First(&installment, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &installment, nil
}

// UpdatePayRollInstallment updates an existing installment for a payroll.
// The installment must exist in the database and have a valid id.
func (s *PayrollService) UpdatePayRollInstallment(installment *models.PayRollInstallment) error {
	return s.db.Save(installment).Error
}

// DeletePayRollInstallment deletes an installment for a payroll with given id.
// The installment must exist in the database before calling this method.
func (s *PayrollService) DeletePayRollInstallment(id string) error {
	return s.db.Delete(&models.PayRollInstallment{}, "id = ?", id).Error
}

// FindAll retrieves a paginated list of payroll installments from the database.
//
// The function takes an HTTP request object as parameter and uses the query
// parameters to filter the payroll installments. The records are sorted by
// created_at in descending order by default, but the order can be changed
// by specifying the "order" query parameter.
//
// The function returns a Page object containing the payroll installments and
// the pagination information. The Page object contains the following fields:
//
//	Records: []models.PayRollInstallment
//	Page: int
//	PageSize: int
//	TotalPages: int
//	TotalRecords: int
//
// If the operation is not successful, the function returns an error object.
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
