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
	err := e.db.First(&employee, id).Error
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
	stmt := e.db.Model(&models.EmployeeModel{})
	utils.FixRequest(request)
	page := pg.With(stmt).Request(&request).Response(&[]models.EmployeeModel{})
	page.Page = page.Page + 1
	return page, nil
}
