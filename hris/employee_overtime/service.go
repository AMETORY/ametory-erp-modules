package employee_overtime

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type EmployeeOvertimeService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewEmployeeOvertimeService(ctx *context.ERPContext) *EmployeeOvertimeService {
	return &EmployeeOvertimeService{db: ctx.DB, ctx: ctx}
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.EmployeeOvertimeModel{},
	)
}

func (e *EmployeeOvertimeService) CreateEmployeeOvertime(employeeOvertime *models.EmployeeOvertimeModel) error {
	return e.db.Create(employeeOvertime).Error
}

func (e *EmployeeOvertimeService) GetEmployeeOvertimeByID(id string) (*models.EmployeeOvertimeModel, error) {
	var employeeOvertime models.EmployeeOvertimeModel
	err := e.db.Where("id = ?", id).First(&employeeOvertime).Error
	if err != nil {
		return nil, err
	}
	return &employeeOvertime, nil
}

func (e *EmployeeOvertimeService) UpdateEmployeeOvertime(employeeOvertime *models.EmployeeOvertimeModel) error {
	return e.db.Save(employeeOvertime).Error
}

func (e *EmployeeOvertimeService) DeleteEmployeeOvertime(id string) error {
	return e.db.Delete(&models.EmployeeOvertimeModel{}, "id = ?", id).Error
}

func (e *EmployeeOvertimeService) FindAllEmployeeOvertimes(request *http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := e.db.Model(&models.EmployeeActivityModel{})
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.EmployeeActivityModel{})
	page.Page = page.Page + 1
	return page, nil
}
