package payroll

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
)

// CreatePayRollCost adds a new payroll cost record to the database.
// It returns an error if the creation fails.
func (s *PayrollService) CreatePayRollCost(cost *models.PayRollCostModel) error {
	return s.db.Create(cost).Error
}

// GetPayRollCostByID retrieves a payroll cost record by ID from the database.
//
// The function takes an ID as input and attempts to fetch the corresponding
// record from the database. It returns the PayRollCost model and an error if the
// retrieval fails. If the record is not found, a nil pointer is returned together
// with a gorm.ErrRecordNotFound error.
func (s *PayrollService) GetPayRollCostByID(id string) (*models.PayRollCostModel, error) {
	var cost models.PayRollCostModel
	err := s.db.First(&cost, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &cost, nil
}

// UpdatePayRollCost updates an existing payroll cost record in the database.
//
// The function takes a PayRollCost model as input and attempts to update
// the corresponding record in the database. It returns an error if the update fails.
func (s *PayrollService) UpdatePayRollCost(cost *models.PayRollCostModel) error {
	return s.db.Save(cost).Error
}

// DeletePayRollCost deletes an existing payroll cost record in the database.
//
// The function takes the ID of the record to be deleted as input and attempts to delete
// the corresponding record in the database. It returns an error if the deletion fails,
// otherwise returns nil.
func (s *PayrollService) DeletePayRollCost(id string) error {
	return s.db.Delete(&models.PayRollCostModel{}, "id = ?", id).Error
}

// FindAllPayRollCosts retrieves a list of all payroll cost records from the database.
//
// The function takes an http request as input and attempts to retrieve the list of
// records according to the request. The request is expected to contain a "page" query
// parameter that specifies the page number of the list to be retrieved. The function
// returns a paginate.Page object that contains the list of records and the total count
// of records in the database. If the retrieval fails, an error is returned.
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
