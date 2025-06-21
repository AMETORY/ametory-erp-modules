package schedule

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/hris/employee"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type ScheduleService struct {
	db              *gorm.DB
	ctx             *context.ERPContext
	employeeService *employee.EmployeeService
}

func NewScheduleService(ctx *context.ERPContext, employeeService *employee.EmployeeService) *ScheduleService {
	return &ScheduleService{db: ctx.DB, ctx: ctx, employeeService: employeeService}
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.ScheduleModel{},
	)
}

func (s *ScheduleService) CreateSchedule(m *models.ScheduleModel) error {

	return s.db.Create(m).Error
}

func (s *ScheduleService) FindAllSchedule(request *http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.Model(&models.ScheduleModel{})
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.ScheduleModel{})
	page.Page = page.Page + 1
	return page, nil
}

func (s *ScheduleService) FindScheduleByID(id string) (*models.ScheduleModel, error) {
	var m models.ScheduleModel
	if err := s.db.Where("id = ?", id).First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (s *ScheduleService) UpdateSchedule(m *models.ScheduleModel) error {
	return s.db.Save(m).Error
}

func (s *ScheduleService) DeleteSchedule(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.ScheduleModel{}).Error
}

func (s *ScheduleService) Delete(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.ScheduleModel{}).Error
}
