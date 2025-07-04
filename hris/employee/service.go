package employee

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type EmployeeService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewEmployeeService(ctx *context.ERPContext) *EmployeeService {
	return &EmployeeService{db: ctx.DB, ctx: ctx}
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.EmployeeModel{},
		&models.JobTitleModel{},
	)
}
func (e *EmployeeService) CreateEmployee(employee *models.EmployeeModel) error {
	return e.db.Create(employee).Error
}

func (e *EmployeeService) GetEmployeeByID(id string) (*models.EmployeeModel, error) {
	var employee models.EmployeeModel
	err := e.db.
		Preload("User").
		Preload("Company").
		Preload("Bank").
		Preload("JobTitle").
		Preload("Branch").
		Preload("WorkLocation").
		Preload("WorkShift").
		First(&employee, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &employee, nil
}

func (e *EmployeeService) UpdateEmployee(employee *models.EmployeeModel) error {
	return e.db.Save(employee).Error
}

func (e *EmployeeService) DeleteEmployee(id string) error {
	return e.db.Delete(&models.EmployeeModel{}, id).Error
}

func (e *EmployeeService) FindAllEmployees(request *http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := e.db.Preload("User").Preload("Company").Preload("JobTitle").Model(&models.EmployeeModel{})
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	if request.URL.Query().Get("search") != "" {
		stmt = stmt.Where("full_name ilike ? or email ilike ? or employee_identity_number ilike ? or address ilike ? or phone ilike ?",
			"%"+request.URL.Query().Get("search")+"%",
			"%"+request.URL.Query().Get("search")+"%",
			"%"+request.URL.Query().Get("search")+"%",
			"%"+request.URL.Query().Get("search")+"%",
			"%"+request.URL.Query().Get("search")+"%",
		)
	}
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.EmployeeModel{})
	page.Page = page.Page + 1
	return page, nil
}

func (e *EmployeeService) GetEmployeeFromUser(userID string, companyID string) (*models.EmployeeModel, error) {
	var employee models.EmployeeModel
	err := e.db.
		First(&employee, "user_id = ? AND company_id = ?", userID, companyID).Error
	if err != nil {
		return nil, err
	}
	return &employee, nil
}
