package payroll

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
)

// CreatePayRollPeriode adds a new payroll period record to the database.
//
// The function takes a PayRollPeriodeModel as input and attempts to create
// a new record in the database. It returns an error if the creation fails,
// otherwise it returns nil.
func (s *PayrollService) CreatePayRollPeriode(periode *models.PayRollPeriodeModel) error {
	return s.db.Create(periode).Error
}

// GetPayRollPeriodeByID returns the payroll periode by given id.
// It includes the related payrolls.
func (s *PayrollService) GetPayRollPeriodeByID(id string) (*models.PayRollPeriodeModel, error) {
	var periode models.PayRollPeriodeModel
	err := s.db.Preload("PayRolls").First(&periode, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &periode, nil
}

// UpdatePayRollPeriode updates an existing payroll period record in the database.
//
// The function takes a PayRollPeriodeModel as input and attempts to update
// the corresponding record in the database. It returns an error if the update fails.
func (s *PayrollService) UpdatePayRollPeriode(periode *models.PayRollPeriodeModel) error {
	return s.db.Save(periode).Error
}

// DeletePayRollPeriode deletes an existing payroll period record in the database.
//
// The function takes the ID of the record to be deleted as input and attempts to delete
// the corresponding record in the database. It returns an error if the deletion fails,
// otherwise returns nil.
func (s *PayrollService) DeletePayRollPeriode(id string) error {
	return s.db.Delete(&models.PayRollPeriodeModel{}, "id = ?", id).Error
}

// FindAllPayRollPeriodes retrieves a paginated list of payroll period records from the database.
//
// The function takes an http.Request as input and applies a filter for the company ID
// if provided in the request header. The result is a paginated page of PayRollPeriodeModel
// records. The function returns an error if the operation fails, otherwise returns
// the paginated page of records.
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
