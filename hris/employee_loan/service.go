package employee_loan

import (
	"errors"
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/hris/employee"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type EmployeeLoanService struct {
	db              *gorm.DB
	ctx             *context.ERPContext
	employeeService *employee.EmployeeService
}

func NewEmployeeLoanService(ctx *context.ERPContext, employeeService *employee.EmployeeService) *EmployeeLoanService {
	return &EmployeeLoanService{db: ctx.DB, ctx: ctx, employeeService: employeeService}
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.EmployeeLoan{},
	)
}

func (s *EmployeeLoanService) CreateEmployeeLoan(m *models.EmployeeLoan) error {
	if m.EmployeeID == nil {
		return errors.New("employee id is required")
	}
	return s.db.Create(m).Error
}

func (s *EmployeeLoanService) FindAllEmployeeLoan(request *http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.Model(&models.EmployeeLoan{})
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.EmployeeLoan{})
	page.Page = page.Page + 1
	return page, nil
}

func (s *EmployeeLoanService) FindEmployeeLoanByID(id string) (*models.EmployeeLoan, error) {
	var m models.EmployeeLoan
	if err := s.db.Where("id = ?", id).First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (s *EmployeeLoanService) UpdateEmployeeLoan(m *models.EmployeeLoan) error {
	return s.db.Save(m).Error
}

func (s *EmployeeLoanService) DeleteEmployeeLoan(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.EmployeeLoan{}).Error
}
