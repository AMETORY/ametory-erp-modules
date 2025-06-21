package reimbursement

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

type ReimbursementService struct {
	db              *gorm.DB
	ctx             *context.ERPContext
	employeeService *employee.EmployeeService
}

func NewReimbursementService(ctx *context.ERPContext, employeeService *employee.EmployeeService) *ReimbursementService {
	return &ReimbursementService{db: ctx.DB, ctx: ctx, employeeService: employeeService}
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.ReimbursementModel{},
	)
}

func (s *ReimbursementService) CreateReimbursement(m *models.ReimbursementModel) error {
	if m.EmployeeID == nil {
		return errors.New("employee id is required")
	}
	return s.db.Create(m).Error
}

func (s *ReimbursementService) FindAllReimbursement(request *http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.Model(&models.ReimbursementModel{})
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.ReimbursementModel{})
	page.Page = page.Page + 1
	return page, nil
}

func (s *ReimbursementService) FindReimbursementByID(id string) (*models.ReimbursementModel, error) {
	var m models.ReimbursementModel
	if err := s.db.Where("id = ?", id).First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (s *ReimbursementService) UpdateReimbursement(m *models.ReimbursementModel) error {
	return s.db.Save(m).Error
}

func (s *ReimbursementService) DeleteReimbursement(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.ReimbursementModel{}).Error
}

func (s *ReimbursementService) Delete(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.ReimbursementModel{}).Error
}
