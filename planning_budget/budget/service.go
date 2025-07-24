package budget

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
)

type BudgetService struct {
	ctx *context.ERPContext
}

// NewBudgetService creates a new instance of BudgetService.
//
// It takes an ERPContext as parameter and returns a pointer to a BudgetService.
func NewBudgetService(ctx *context.ERPContext) *BudgetService {
	return &BudgetService{ctx: ctx}
}

// CreateBudget creates a new budget in the database.
//
// It takes a pointer to a BudgetModel as parameter and returns the same pointer
// and an error. The error is returned from the GORM Create method.
//
// The returned BudgetModel is the same as the parameter, but with the ID set.
func (s *BudgetService) CreateBudget(model *models.BudgetModel) (*models.BudgetModel, error) {
	return model, s.ctx.DB.Create(&model).Error
}

// GetBudgetByID retrieves a budget by its ID from the database.
//
// It takes an ID string as a parameter and returns a pointer to the BudgetModel
// and an error if the operation was unsuccessful. The function uses the GORM
// library to perform the database query.
func (s *BudgetService) GetBudgetByID(id string) (*models.BudgetModel, error) {
	model := models.BudgetModel{}
	err := s.ctx.DB.
		Where("id = ?", id).
		First(&model).
		Error
	return &model, err
}

// GetBudgets retrieves a paginated list of budgets from the database.
//
// It takes an HTTP request and a search query string as input. The search query
// is applied to the budget name and description fields. If the request contains a
// company ID header, the result is filtered by the company ID. The function uses
// pagination to manage the result set and applies any necessary request
// modifications using the utils.FixRequest utility.
//
// The function returns a paginated page of BudgetModel and an error if the
// operation fails.
func (s *BudgetService) GetBudgets(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.ctx.DB
	if search != "" {
		stmt = stmt.Where("name ILIKE ? OR  description ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
		)
	}
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ? or company_id is null", request.Header.Get("ID-Company"))
	}
	request.URL.Query().Get("page")
	stmt = stmt.Model(&models.UnitModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.UnitModel{})
	page.Page = page.Page + 1
	return page, nil
}

// UpdateBudget updates an existing budget in the database.
//
// It takes a pointer to a BudgetModel as input, which contains the updated budget data.
// The method uses the GORM library to update the budget record where the ID matches the
// provided model's ID. It returns an error if the update operation fails.
func (s *BudgetService) UpdateBudget(model *models.BudgetModel) error {
	return s.ctx.DB.Model(&model).Where("id = ?", model.ID).Updates(model).Error

}

// DeleteBudget deletes a budget from the database by its ID.
//
// It takes an ID string as a parameter and attempts to delete the corresponding
// BudgetModel record from the database. The method returns an error if the
// deletion operation fails.
func (s *BudgetService) DeleteBudget(id string) error {
	return s.ctx.DB.Delete(&models.BudgetModel{}, "id = ?", id).Error
}
