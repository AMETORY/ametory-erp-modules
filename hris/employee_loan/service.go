package employee_loan

import (
	"errors"
	"net/http"
	"time"

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

// NewEmployeeLoanService creates a new EmployeeLoanService instance.
//
// The EmployeeLoanService is used to manage employee loans, which are
// financial assistance given to employees by the company.
//
// The service requires a pointer to an ERPContext, which contains
// the user's HTTP request context and other relevant information.
// The service also requires a pointer to an EmployeeService, which is
// used to fetch related employee data.
//
// The service will call the Migrate() method after creation to migrate
// the database. If the migration fails, the service will panic.
func NewEmployeeLoanService(ctx *context.ERPContext, employeeService *employee.EmployeeService) *EmployeeLoanService {
	return &EmployeeLoanService{db: ctx.DB, ctx: ctx, employeeService: employeeService}
}

// Migrate migrates the EmployeeLoan database table.
//
// The function takes a pointer to a Gorm DB as argument and returns an error.
// The function will panic if the migration fails.
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.EmployeeLoan{},
	)
}

// CreateEmployeeLoan creates a new employee loan record in the database.
//
// The function takes an EmployeeLoan model as input and ensures that
// the EmployeeID is present before attempting to create the record.
// If the EmployeeID is nil, an error is returned indicating that the
// employee ID is required. If the creation is successful, the function
// returns nil, otherwise it returns an error.
func (s *EmployeeLoanService) CreateEmployeeLoan(m *models.EmployeeLoan) error {
	if m.EmployeeID == nil {
		return errors.New("employee id is required")
	}
	return s.db.Create(m).Error
}

// FindAllByEmployeeID retrieves a paginated list of employee loans for a given employee ID.
//
// This function takes an HTTP request and an employee ID as parameters.
// It uses GORM to query the database for employee loans, preloading the associated
// Employee (with User, JobTitle, WorkLocation, WorkShift, Branch), Company, and Approver models.
// The function applies various filters based on the request parameters, such as company ID,
// search term, date range, specific date, and approver ID.
// The results can be ordered based on a specified order parameter or defaults to ordering by date in descending order.
//
// Pagination is utilized to manage the result set, and any necessary request modifications
// are applied using the utils.FixRequest utility. The function returns a paginated page of
// EmployeeLoan and an error if the operation fails.
func (s *EmployeeLoanService) FindAllByEmployeeID(request *http.Request, employeeID string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.
		Preload("Employee", func(tx *gorm.DB) *gorm.DB {
			return tx.Preload("User").
				Preload("JobTitle").
				Preload("WorkLocation").
				Preload("WorkShift").
				Preload("Branch")
		}).
		Preload("Company").
		Preload("Approver").
		Model(&models.EmployeeLoan{}).Where("employee_id = ?", employeeID)
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	if request.URL.Query().Get("search") != "" {
		stmt = stmt.Where("reason LIKE ?", "%"+request.URL.Query().Get("search")+"%")
	}
	if request.URL.Query().Get("start_date") != "" && request.URL.Query().Get("end_date") != "" {
		stmt = stmt.Where("date >= ? AND date <= ?", request.URL.Query().Get("start_date"), request.URL.Query().Get("end_date"))
	} else if request.URL.Query().Get("start_date") != "" {
		stmt = stmt.Where("date = ?", request.URL.Query().Get("start_date"))
	}
	if request.URL.Query().Get("date") != "" {
		stmt = stmt.Where("DATE(date) = ?", request.URL.Query().Get("date"))
	}
	if request.URL.Query().Get("approver_id") != "" {
		stmt = stmt.Where("approver_id = ?", request.URL.Query().Get("approver_id"))
	}
	if request.URL.Query().Get("order") != "" {
		stmt = stmt.Order(request.URL.Query().Get("order"))
	} else {
		stmt = stmt.Order("date DESC")
	}
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.EmployeeLoan{})
	page.Page = page.Page + 1
	return page, nil
}

// FindAllEmployeeLoan retrieves a paginated list of all employee loans.
//
// The function uses GORM to query the database for employee loans, preloading the
// associated Employee (with User, JobTitle, WorkLocation, WorkShift, Branch), Company,
// and Approver models. It applies various filters based on the request parameters such as
// company ID, search term, date range, specific date, and approver ID. The results can be
// ordered based on a specified order parameter or defaults to ordering by date in descending order.
//
// Pagination is utilized to manage the result set, and any necessary request modifications
// are applied using the utils.FixRequest utility. The function returns a paginated page of
// EmployeeLoan and an error if the operation fails.
func (s *EmployeeLoanService) FindAllEmployeeLoan(request *http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.
		Preload("Employee", func(tx *gorm.DB) *gorm.DB {
			return tx.Preload("User").
				Preload("JobTitle").
				Preload("WorkLocation").
				Preload("WorkShift").
				Preload("Branch")
		}).
		Preload("Company").
		Preload("Approver.User").
		Model(&models.EmployeeLoan{})
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	if request.URL.Query().Get("search") != "" {
		stmt = stmt.Where("reason LIKE ?", "%"+request.URL.Query().Get("search")+"%")
	}
	if request.URL.Query().Get("start_date") != "" && request.URL.Query().Get("end_date") != "" {
		stmt = stmt.Where("date >= ? AND date <= ?", request.URL.Query().Get("start_date"), request.URL.Query().Get("end_date"))
	} else if request.URL.Query().Get("start_date") != "" {
		stmt = stmt.Where("date = ?", request.URL.Query().Get("start_date"))
	}
	if request.URL.Query().Get("date") != "" {
		stmt = stmt.Where("DATE(date) = ?", request.URL.Query().Get("date"))
	}
	if request.URL.Query().Get("approver_id") != "" {
		stmt = stmt.Where("approver_id = ?", request.URL.Query().Get("approver_id"))
	}
	if request.URL.Query().Get("order") != "" {
		stmt = stmt.Order(request.URL.Query().Get("order"))
	} else {
		stmt = stmt.Order("date DESC")
	}
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.EmployeeLoan{})
	page.Page = page.Page + 1
	return page, nil
}

// FindEmployeeLoanByID retrieves an employee loan by its ID from the database.
//
// The function preloads the associated Employee (with User, JobTitle, WorkLocation, WorkShift, Branch),
// Approver (with User), and ApprovalByAdmin models. It also attempts to find an associated FileModel
// based on the loan's ID, setting it if found. Returns the populated EmployeeLoan model and any error encountered.
func (s *EmployeeLoanService) FindEmployeeLoanByID(id string) (*models.EmployeeLoan, error) {
	var m models.EmployeeLoan
	if err := s.db.
		Preload("Employee", func(tx *gorm.DB) *gorm.DB {
			return tx.Preload("User").
				Preload("JobTitle").
				Preload("WorkLocation").
				Preload("WorkShift").
				Preload("Branch")
		}).
		Preload("Approver.User").
		Preload("ApprovalByAdmin").
		Where("id = ?", id).First(&m).Error; err != nil {
		return nil, err
	}

	file := models.FileModel{}
	s.ctx.DB.Find(&file, "ref_id = ? AND ref_type = ?", id, "employee_loan")
	if file.ID != "" {
		m.File = &file
	}
	return &m, nil
}

// UpdateEmployeeLoan updates an existing employee loan record in the database.
//
// The function takes an EmployeeLoan model as input and attempts to update
// the corresponding record in the database. It returns an error if the update
// fails, otherwise returns nil.
func (s *EmployeeLoanService) UpdateEmployeeLoan(m *models.EmployeeLoan) error {
	return s.db.Save(m).Error
}

func (s *EmployeeLoanService) DeleteEmployeeLoan(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.EmployeeLoan{}).Error
}

// CountByEmployeeID returns a map of counts of employee loan records
// filtered by a given employee ID and date range, and grouped by status.
//
// The map has the following keys:
// - REQUESTED: the count of records with "REQUESTED" status
// - APPROVED: the count of records with "APPROVED" status
// - REJECTED: the count of records with "REJECTED" status
//
// Parameters:
//   - employeeID: The ID of the employee whose loan records are being counted.
//   - startDate: The start date of the date range for filtering loan records.
//   - endDate: The end date of the date range for filtering loan records.
//
// Returns:
//   - A map of loan counts by status.
//   - An error if the operation fails.
func (e *EmployeeLoanService) CountByEmployeeID(employeeID string, startDate *time.Time, endDate *time.Time) (map[string]int64, error) {
	var countREQUESTED, countAPPROVED, countREJECTED int64
	counts := make(map[string]int64)
	e.db.Model(&models.EmployeeLoan{}).
		Where("employee_id = ? AND status = ? AND date >= ? AND date <= ?", employeeID, "REQUESTED", startDate, endDate).
		Count(&countREQUESTED)
	e.db.Model(&models.EmployeeLoan{}).
		Where("employee_id = ? AND status = ? AND date >= ? AND date <= ?", employeeID, "APPROVED", startDate, endDate).
		Count(&countAPPROVED)
	e.db.Model(&models.EmployeeLoan{}).
		Where("employee_id = ? AND status = ? AND date >= ? AND date <= ?", employeeID, "REJECTED", startDate, endDate).
		Count(&countREJECTED)

	counts["REQUESTED"] = countREQUESTED
	counts["APPROVED"] = countAPPROVED
	counts["REJECTED"] = countREJECTED

	return counts, nil
}
