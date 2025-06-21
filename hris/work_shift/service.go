package work_shift

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/hris/employee"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type WorkShiftService struct {
	db              *gorm.DB
	ctx             *context.ERPContext
	employeeService *employee.EmployeeService
}

func NewWorkShiftService(ctx *context.ERPContext, employeeService *employee.EmployeeService) *WorkShiftService {
	return &WorkShiftService{db: ctx.DB, ctx: ctx, employeeService: employeeService}
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.WorkShiftModel{},
	)
}

func (a *WorkShiftService) CreateWorkShift(m *models.WorkShiftModel) error {
	return a.db.Create(m).Error
}

func (a *WorkShiftService) FindWorkShiftByID(id string) (*models.WorkShiftModel, error) {
	m := &models.WorkShiftModel{}
	if err := a.db.Where("id = ?", id).First(m).Error; err != nil {
		return nil, err
	}

	return m, nil
}

func (a *WorkShiftService) FindAllWorkShift(request *http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := a.db.Model(&models.WorkShiftModel{})
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.WorkShiftModel{})
	page.Page = page.Page + 1
	return page, nil
}

func (a *WorkShiftService) UpdateWorkShift(id string, m *models.WorkShiftModel) error {
	return a.db.Where("id = ?", id).Updates(m).Error
}

func (a *WorkShiftService) DeleteWorkShift(id string) error {
	return a.db.Where("id = ?", id).Delete(&models.WorkShiftModel{}).Error
}
