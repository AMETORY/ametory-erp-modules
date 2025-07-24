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

func NewBudgetService(ctx *context.ERPContext) *BudgetService {
	return &BudgetService{ctx: ctx}
}

func (s *BudgetService) CreateBudget(model *models.BudgetModel) (*models.BudgetModel, error) {
	return model, s.ctx.DB.Create(&model).Error
}

func (s *BudgetService) GetBudgetByID(id string) (*models.BudgetModel, error) {
	model := models.BudgetModel{}
	err := s.ctx.DB.
		Where("id = ?", id).
		First(&model).
		Error
	return &model, err
}

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

func (s *BudgetService) UpdateBudget(model *models.BudgetModel) error {
	return s.ctx.DB.Model(&model).Where("id = ?", model.ID).Updates(model).Error

}

func (s *BudgetService) DeleteBudget(id string) error {
	return s.ctx.DB.Delete(&models.BudgetModel{}, "id = ?", id).Error
}
