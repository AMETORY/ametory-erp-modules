package schedule

import (
	"net/http"
	"time"

	"slices"

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

func (r *ScheduleService) FindApplicableSchedulesForEmployee(employee *models.EmployeeModel, now time.Time) ([]models.ScheduleModel, error) {
	var schedules []models.ScheduleModel

	err := r.db.
		Joins("LEFT JOIN schedule_employees ON schedule_employees.schedule_model_id = schedules.id").
		Joins("LEFT JOIN schedule_organizations ON schedule_organizations.schedule_model_id = schedules.id").
		Joins("LEFT JOIN schedule_branches ON schedule_branches.schedule_model_id = schedules.id").
		Where("schedules.is_active = true").
		Where(`
			schedule_employees.employee_model_id = ? OR
			(schedule_organizations.organization_model_id = ? OR schedule_organizations.organization_model_id IS NULL) OR
			(schedule_branches.branch_model_id = ? OR schedule_branches.branch_model_id IS NULL)`,
			employee.ID,
			utils.StringOrEmpty(employee.OrganizationID),
			utils.StringOrEmpty(employee.BranchID),
		).
		Where("(schedule_employees.effective_date IS NULL OR schedule_employees.effective_date <= ?) AND (schedule_employees.effective_until IS NULL OR schedule_employees.effective_until >= ?)",
			time.Now(), time.Now()).
		Group("schedules.id").
		Find(&schedules).Error

	if err != nil {
		return nil, err
	}

	// Filter ulang berdasarkan tanggal dan repeat type
	var result []models.ScheduleModel
	for _, sched := range schedules {
		if r.IsScheduleApplicable(&sched, now, employee) {
			result = append(result, sched)
		}
	}

	return result, nil
}

func (r *ScheduleService) IsScheduleApplicable(s *models.ScheduleModel, t time.Time, employee *models.EmployeeModel) bool {
	// 1. Validasi tanggal aktif
	if t.Before(*s.StartDate) || (s.EndDate != nil && t.After(*s.EndDate)) {
		return false
	}

	// 2. Cek apakah schedule ini memang berlaku untuk employee ini
	if !r.isScheduleAssignedToEmployee(s, employee) {
		return false
	}

	// 3. Validasi pola pengulangan
	switch s.RepeatType {
	case "ONCE":
		return s.StartDate.Format("2006-01-02") == t.Format("2006-01-02")
	case "DAILY":
		return true
	case "WEEKLY":
		day := t.Weekday().String() // e.g. "Monday"
		return slices.Contains(s.RepeatDays, day)
	default:
		return false
	}
}

func (r *ScheduleService) isScheduleAssignedToEmployee(s *models.ScheduleModel, e *models.EmployeeModel) bool {
	// Prioritas paling tinggi: langsung diassign ke employee
	for _, emp := range s.Employees {
		if emp.ID == e.ID {
			return true
		}
	}

	// Cek organisasi
	if e.OrganizationID != nil {
		for _, org := range s.Organizations {
			if org.ID == *e.OrganizationID {
				return true
			}
		}
	}

	// Cek branch
	if e.BranchID != nil {
		for _, branch := range s.Branches {
			if branch.ID == *e.BranchID {
				return true
			}
		}
	}

	return false
}
