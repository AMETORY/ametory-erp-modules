package leave

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

type LeaveService struct {
	db              *gorm.DB
	ctx             *context.ERPContext
	employeeService *employee.EmployeeService
}

func NewLeaveService(ctx *context.ERPContext, employeeService *employee.EmployeeService) *LeaveService {
	return &LeaveService{db: ctx.DB, ctx: ctx, employeeService: employeeService}
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.LeaveModel{},
		&models.LeaveCategory{},
	)
}

func (s *LeaveService) CreateLeave(m *models.LeaveModel) error {
	if m.EmployeeID == nil {
		return errors.New("employee id is required")
	}
	return s.db.Create(m).Error
}

func (s *LeaveService) FindAllLeave(request *http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.Preload("LeaveCategory").Model(&models.LeaveModel{})
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.LeaveModel{})
	page.Page = page.Page + 1
	return page, nil
}

func (s *LeaveService) FindLeaveByID(id string) (*models.LeaveModel, error) {
	var m models.LeaveModel
	if err := s.db.Preload("LeaveCategory").Where("id = ?", id).First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (s *LeaveService) UpdateLeave(m *models.LeaveModel) error {
	return s.db.Save(m).Error
}

func (s *LeaveService) DeleteLeave(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.LeaveModel{}).Error
}

func (s *LeaveService) Delete(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.LeaveModel{}).Error
}

func (s *LeaveService) CreateLeaveCategory(c *models.LeaveCategory) error {
	return s.db.Create(c).Error
}

func (s *LeaveService) FindAllLeaveCategories(request *http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.Model(&models.LeaveCategory{})
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.LeaveCategory{})
	page.Page = page.Page + 1
	return page, nil
}

func (s *LeaveService) FindLeaveCategoryByID(id string) (*models.LeaveCategory, error) {
	var category models.LeaveCategory
	if err := s.db.Where("id = ?", id).First(&category).Error; err != nil {
		return nil, err
	}
	return &category, nil
}

func (s *LeaveService) UpdateLeaveCategory(c *models.LeaveCategory) error {
	return s.db.Save(c).Error
}

func (s *LeaveService) DeleteLeaveCategory(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.LeaveCategory{}).Error
}
