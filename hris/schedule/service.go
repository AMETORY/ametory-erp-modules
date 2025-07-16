package schedule

import (
	"fmt"
	"net/http"
	"strings"
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

// NewScheduleService creates a new instance of ScheduleService.
//
// The service is created by providing a pointer to an ERPContext and an
// EmployeeService. The ERPContext is used for authentication and authorization
// purposes, while the EmployeeService is used to fetch related data.
func NewScheduleService(ctx *context.ERPContext, employeeService *employee.EmployeeService) *ScheduleService {
	return &ScheduleService{db: ctx.DB, ctx: ctx, employeeService: employeeService}
}

// Migrate runs the auto-migration for the ScheduleModel database table.
//
// It takes a pointer to a GORM DB as input and returns an error if the migration
// fails. The migration creates the ScheduleModel table if it does not exist,
// and updates the existing table if it does.
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.ScheduleModel{},
	)
}

// CreateSchedule creates a new schedule entry in the database.
//
// It takes a pointer to a ScheduleModel as input and returns an error if the
// creation fails. The ScheduleModel must have a valid UUID as its ID field;
// otherwise, the creation will fail.
func (s *ScheduleService) CreateSchedule(m *models.ScheduleModel) error {

	return s.db.Create(m).Error
}

// FindAllSchedule retrieves a paginated list of schedule records from the database.
//
// The function takes an http request as input and attempts to retrieve the list of
// records according to the request. The request is expected to contain a "page" query
// parameter that specifies the page number of the list to be retrieved. The function
// returns a paginate.Page object that contains the list of records and the total count
// of records in the database. If the retrieval fails, an error is returned.
func (s *ScheduleService) FindAllSchedule(request *http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.Preload("Branches").Preload("Employees").Preload("Organizations").Model(&models.ScheduleModel{})
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.ScheduleModel{})
	page.Page = page.Page + 1
	return page, nil
}

// FindScheduleByID retrieves a schedule record by ID from the database.
//
// The function takes a schedule ID as input and uses it to query the database for the
// schedule record. The associated Branches, Employees, and Organizations models are
// preloaded.
//
// The function returns the schedule record and an error if the operation fails. If the
// schedule record is not found, a nil pointer is returned together with a gorm.ErrRecordNotFound
// error.
func (s *ScheduleService) FindScheduleByID(id string) (*models.ScheduleModel, error) {
	var m models.ScheduleModel
	if err := s.db.Where("id = ?", id).First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

// UpdateSchedule updates an existing schedule record in the database.
//
// It takes a pointer to a ScheduleModel as input and attempts to update the
// corresponding record in the database. If the update operation is successful,
// it returns nil; otherwise, it returns an error.
//
// Note that the ScheduleModel must have a valid UUID as its ID field; otherwise,
// the update operation will fail.
func (s *ScheduleService) UpdateSchedule(m *models.ScheduleModel) error {
	return s.db.Save(m).Error
}

// DeleteSchedule deletes an existing schedule record in the database.
//
// The function takes a schedule ID as parameter and attempts to delete the
// schedule record with the given ID from the database. If the operation is
// successful, it returns nil; otherwise, it returns an error.
func (s *ScheduleService) DeleteSchedule(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.ScheduleModel{}).Error
}

// Delete deletes an existing schedule record in the database.
//
// The function takes a schedule ID as parameter and attempts to delete the
// schedule record with the given ID from the database. If the operation is
// successful, it returns nil; otherwise, it returns an error.
func (s *ScheduleService) Delete(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.ScheduleModel{}).Error
}

// FindApplicableSchedulesForEmployee finds applicable schedules for the given employee
// at the given time, considering the employee's organization and branch.
//
// The function takes an employee model and a time as input and returns a list of
// schedule models that are applicable to the employee at the given time. The
// applicable schedules are filtered by organization, branch, effective date, and
// repeat type.
//
// If the operation fails, an error is returned.
func (r *ScheduleService) FindApplicableSchedulesForEmployee(employee *models.EmployeeModel, now time.Time) ([]models.ScheduleModel, error) {
	var schedules []models.ScheduleModel = []models.ScheduleModel{}

	err := r.db.Preload("Branches").Preload("Employees").Preload("Organizations").
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
		Where("(schedules.effective_date IS NULL OR schedules.effective_date <= ?) AND (schedules.effective_until IS NULL OR schedules.effective_until >= ?)",
			time.Now(), time.Now()).
		Group("schedules.id").
		Find(&schedules).Error

	if err != nil {
		return nil, err
	}

	// Filter ulang berdasarkan tanggal dan repeat type
	var result []models.ScheduleModel = []models.ScheduleModel{}
	for _, sched := range schedules {
		if r.IsScheduleApplicable(&sched, now, employee) {
			result = append(result, sched)
		}
	}

	return result, nil
}

// IsScheduleApplicable checks whether a schedule is applicable to a given employee at a given time.
//
// The function takes a schedule model, a time, and an employee model as input and checks
// whether the schedule is applicable to the employee at the given time. The function
// checks the following conditions:
//  1. The schedule is active (i.e. effective date is before or equal to the given time
//     and the effective until date is after or equal to the given time).
//  2. The schedule is assigned to the given employee (either directly or through the
//     employee's organization or branch).
//  3. The schedule's repeat type is valid for the given time (e.g. if the repeat type
//     is "DAILY", the function checks whether the given time is within the schedule's
//     start and end dates; if the repeat type is "WEEKLY", the function checks whether
//     the given time is within the schedule's start and end dates and whether the day
//     of the week of the given time matches one of the repeat days).
//
// If all conditions are met, the function returns true; otherwise, it returns false.
func (r *ScheduleService) IsScheduleApplicable(s *models.ScheduleModel, t time.Time, employee *models.EmployeeModel) bool {
	// 1. Validasi tanggal aktif
	if (s.EffectiveDate != nil && t.Before(*s.EffectiveDate)) || (s.EffectiveUntil != nil && t.After(*s.EffectiveUntil)) {
		return false
	}

	fmt.Println(s.RepeatType)

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
		return slices.Contains(s.RepeatDays, strings.ToUpper(day))
	default:
		return false
	}
}

// isScheduleAssignedToEmployee checks whether a schedule is assigned to a given employee.
//
// The function takes a schedule model and an employee model as input and checks
// whether the schedule is assigned to the employee either directly, through the
// employee's organization, or through the employee's branch. The function returns
// true if the schedule is assigned to the employee; otherwise, it returns false.
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

	// // Cek workshift
	// if e.WorkShiftID != nil {
	// 	for _, ws := range s.WorkShifts {
	// 		if ws.ID == *e.WorkShiftID {
	// 			return true
	// 		}
	// 	}
	// }

	return false
}
