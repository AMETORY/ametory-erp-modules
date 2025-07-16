package employee_cash_advance

import (
	"net/http"
	"strings"
	"time"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type EmployeeCashAdvanceService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

// NewEmployeeCashAdvanceService creates a new instance of EmployeeCashAdvanceService with the provided ERPContext.
//
// The ERPContext is used for authentication and authorization purposes.
func NewEmployeeCashAdvanceService(ctx *context.ERPContext) *EmployeeCashAdvanceService {
	return &EmployeeCashAdvanceService{db: ctx.DB, ctx: ctx}
}

// Migrate runs the auto migration for the EmployeeCashAdvance and related models.
//
// Auto migration will create the necessary table in the database if it does not
// exist, and will also add any missing columns. It will not delete any existing
// columns or data.
//
// It is recommended to call this function once when the application starts, and
// after any changes to the models are made.
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.EmployeeCashAdvance{},
		&models.CashAdvanceUsage{},
		&models.CashAdvanceRefund{},
	)
}

// CreateEmployeeCashAdvance adds a new employee cash advance record to the database.
//
// The function takes an EmployeeCashAdvance model as input and attempts
// to create a new record in the database. It returns an error if the
// creation fails, otherwise returns nil.
func (e *EmployeeCashAdvanceService) CreateEmployeeCashAdvance(employeeCashAdvance *models.EmployeeCashAdvance) error {
	return e.db.Create(employeeCashAdvance).Error
}

// GetEmployeeCashAdvanceByID retrieves an employee cash advance by ID from the database.
//
// The function takes an ID as input and attempts to fetch the corresponding
// record from the database. It returns the EmployeeCashAdvance model and an
// error if the retrieval fails. If the record is not found, a nil pointer is
// returned together with a gorm.ErrRecordNotFound error.
//
// The function also preloads the following related models:
//
// - Employee with User, JobTitle, WorkLocation, WorkShift, and Branch
// - Approver with User
// - ApprovalByAdmin
// - RefundApprovalByAdmin
// - CashAdvanceUsages
// - Refunds
//
// Additionally, it also fetches the file records associated with the
// employee cash advance and its related models, and stores them in the
// respective models.
func (e *EmployeeCashAdvanceService) GetEmployeeCashAdvanceByID(id string) (*models.EmployeeCashAdvance, error) {
	var employeeCashAdvance models.EmployeeCashAdvance
	err := e.db.
		Preload("Employee", func(tx *gorm.DB) *gorm.DB {
			return tx.Preload("User").
				Preload("JobTitle").
				Preload("WorkLocation").
				Preload("WorkShift").
				Preload("Branch")
		}).
		Preload("Approver.User").
		Preload("ApprovalByAdmin").
		Preload("RefundApprovalByAdmin").
		Preload("CashAdvanceUsages").
		Preload("Refunds").
		Where("id = ?", id).First(&employeeCashAdvance).Error
	if err != nil {
		return nil, err
	}

	file := models.FileModel{}
	e.ctx.DB.Find(&file, "ref_id = ? AND ref_type = ?", id, "employee_cash_advance")
	if file.ID != "" {
		employeeCashAdvance.File = &file
	}

	for i, v := range employeeCashAdvance.CashAdvanceUsages {
		files := []models.FileModel{}
		e.db.Find(&files, "ref_id = ? AND ref_type = ?", v.ID, "cash_advance_usage")
		v.Files = files
		employeeCashAdvance.CashAdvanceUsages[i] = v
	}
	for i, v := range employeeCashAdvance.Refunds {
		files := []models.FileModel{}
		e.db.Find(&files, "ref_id = ? AND ref_type = ?", v.ID, "cash_advance_refund")
		v.Files = files
		employeeCashAdvance.Refunds[i] = v
	}

	return &employeeCashAdvance, nil
}

// UpdateEmployeeCashAdvance updates an existing employee cash advance record in the database.
//
// The function takes an EmployeeCashAdvance model as input and attempts to update
// the corresponding record in the database. It returns an error if the update
// fails, otherwise returns nil.
func (e *EmployeeCashAdvanceService) UpdateEmployeeCashAdvance(employeeCashAdvance *models.EmployeeCashAdvance) error {
	return e.db.Save(employeeCashAdvance).Error
}

// DeleteEmployeeCashAdvance deletes an existing employee cash advance record in the database.
//
// The function takes an ID as input and attempts to delete the corresponding
// record in the database. It returns an error if the deletion fails, otherwise
// returns nil.
func (e *EmployeeCashAdvanceService) DeleteEmployeeCashAdvance(id string) error {
	return e.db.Delete(&models.EmployeeCashAdvance{}, "id = ?", id).Error
}

// FindAllByEmployeeID retrieves a paginated list of employee cash advances by employee ID.
//
// The function uses GORM to query the database for employee cash advances, preloading the
// associated Employee (with User, JobTitle, WorkLocation, WorkShift, Branch) and Approver (with User)
// models. It applies various filters based on the request parameters such as company ID,
// search term, date range, specific date, and approver ID. The results can be ordered based on
// a specified order parameter or defaults to ordering by date_requested in descending order.
//
// It utilizes pagination to manage the result set and applies any necessary request modifications
// using the utils.FixRequest utility. The function returns a paginated page of EmployeeCashAdvance
// and an error if the operation fails.

func (e *EmployeeCashAdvanceService) FindAllByEmployeeID(request *http.Request, employeeID string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := e.db.
		Preload("Employee", func(tx *gorm.DB) *gorm.DB {
			return tx.Preload("User").
				Preload("JobTitle").
				Preload("WorkLocation").
				Preload("WorkShift").
				Preload("Branch")
		}).
		Preload("Approver.User").
		Preload("ApprovalByAdmin").
		Where("employee_id = ?", employeeID).
		Model(&models.EmployeeCashAdvance{})
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}

	if request.URL.Query().Get("search") != "" {
		stmt = stmt.Where("reason LIKE ?", "%"+request.URL.Query().Get("search")+"%")
	}
	if request.URL.Query().Get("start_date") != "" && request.URL.Query().Get("end_date") != "" {
		stmt = stmt.Where("date_requested >= ? AND date_requested <= ?", request.URL.Query().Get("start_date"), request.URL.Query().Get("end_date"))
	} else if request.URL.Query().Get("start_date") != "" {
		stmt = stmt.Where("date_requested = ?", request.URL.Query().Get("start_date"))
	}
	if request.URL.Query().Get("date") != "" {
		stmt = stmt.Where("DATE(date_requested) = ?", request.URL.Query().Get("date"))
	}
	if request.URL.Query().Get("approver_id") != "" {
		stmt = stmt.Where("approver_id = ?", request.URL.Query().Get("approver_id"))
	}

	if request.URL.Query().Get("order") != "" {
		stmt = stmt.Order(request.URL.Query().Get("order"))
	} else {
		stmt = stmt.Order("date_requested DESC")
	}
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.EmployeeCashAdvance{})
	page.Page = page.Page + 1
	return page, nil
}

// FindAllEmployeeCashAdvances retrieves a paginated list of employee cash advances.
//
// The function uses GORM to query the database for employee cash advances, preloading the
// associated Employee (with User, JobTitle, WorkLocation, WorkShift, Branch) and Approver (with User)
// models. It applies various filters based on the request parameters such as company ID,
// search term, date range, specific date, and approver ID. The results can be ordered based on
// a specified order parameter or defaults to ordering by date_requested in descending order.
//
// It utilizes pagination to manage the result set and applies any necessary request modifications
// using the utils.FixRequest utility. The function returns a paginated page of EmployeeCashAdvance
// and an error if the operation fails.
func (e *EmployeeCashAdvanceService) FindAllEmployeeCashAdvances(request *http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := e.db.
		Preload("Employee", func(tx *gorm.DB) *gorm.DB {
			return tx.Preload("User").
				Preload("JobTitle").
				Preload("WorkLocation").
				Preload("WorkShift").
				Preload("Branch")
		}).
		Preload("Approver.User").
		Preload("ApprovalByAdmin").
		Model(&models.EmployeeCashAdvance{})
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}

	if request.URL.Query().Get("search") != "" {
		stmt = stmt.Where("reason LIKE ?", "%"+request.URL.Query().Get("search")+"%")
	}
	if request.URL.Query().Get("start_date") != "" && request.URL.Query().Get("end_date") != "" {
		stmt = stmt.Where("date_requested >= ? AND date_requested <= ?", request.URL.Query().Get("start_date"), request.URL.Query().Get("end_date"))
	} else if request.URL.Query().Get("start_date") != "" {
		stmt = stmt.Where("date_requested = ?", request.URL.Query().Get("start_date"))
	}
	if request.URL.Query().Get("date") != "" {
		stmt = stmt.Where("DATE(date_requested) = ?", request.URL.Query().Get("date"))
	}
	if request.URL.Query().Get("approver_id") != "" {
		stmt = stmt.Where("approver_id = ?", request.URL.Query().Get("approver_id"))
	}

	if request.URL.Query().Get("order") != "" {
		stmt = stmt.Order(request.URL.Query().Get("order"))
	} else {
		stmt = stmt.Order("date_requested DESC")
	}

	if request.URL.Query().Get("start_date") != "" {
		stmt = stmt.Where("start_date >= ?", request.URL.Query().Get("start_date"))
	}

	if request.URL.Query().Get("end_date") != "" {
		stmt = stmt.Where("end_date <= ?", request.URL.Query().Get("end_date"))
	}

	if request.URL.Query().Get("employee_ids") != "" {
		stmt = stmt.Where("employee_id in (?)", strings.Split(request.URL.Query().Get("employee_ids"), ","))
	}
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.EmployeeCashAdvance{})
	page.Page = page.Page + 1
	return page, nil
}

// CountByEmployeeID returns a map of counts of employee cash advance records
// filtered by given employee ID and date range, and grouped by status.
//
// The map has the following keys:
// - REQUESTED: the count of records with "REQUESTED" status
// - APPROVED: the count of records with "APPROVED" status
// - REJECTED: the count of records with "REJECTED" status
func (e *EmployeeCashAdvanceService) CountByEmployeeID(employeeID string, startDate *time.Time, endDate *time.Time) (map[string]int64, error) {
	var countREQUESTED, countAPPROVED, countREJECTED int64
	counts := make(map[string]int64)
	e.db.Model(&models.EmployeeCashAdvance{}).
		Where("employee_id = ? AND status = ? AND date_requested >= ? AND date_requested <= ?", employeeID, "REQUESTED", startDate, endDate).
		Count(&countREQUESTED)
	e.db.Model(&models.EmployeeCashAdvance{}).
		Where("employee_id = ? AND status = ? AND date_requested >= ? AND date_requested <= ?", employeeID, "APPROVED", startDate, endDate).
		Count(&countAPPROVED)
	e.db.Model(&models.EmployeeCashAdvance{}).
		Where("employee_id = ? AND status = ? AND date_requested >= ? AND date_requested <= ?", employeeID, "REJECTED", startDate, endDate).
		Count(&countREJECTED)

	counts["REQUESTED"] = countREQUESTED
	counts["APPROVED"] = countAPPROVED
	counts["REJECTED"] = countREJECTED

	return counts, nil
}

// CreateCashAdvanceUsage adds a new cash advance usage record to the database.
//
// The function takes a CashAdvanceUsage model as input and attempts to create
// a new record in the database. It returns an error if the creation fails,
// otherwise returns nil.
func (e *EmployeeCashAdvanceService) CreateCashAdvanceUsage(cashAdvanceUsage *models.CashAdvanceUsage) error {
	return e.db.Create(cashAdvanceUsage).Error
}

// UpdateEmployeeCashAdvanceUsage updates an existing employee cash advance usage record in the database.
//
// The function takes the ID of the record to be updated and a pointer to a CashAdvanceUsage model as input, and
// attempts to update the corresponding record in the database. It returns an error if the update fails, otherwise
// returns nil.
func (e *EmployeeCashAdvanceService) UpdateEmployeeCashAdvanceUsage(id string, input *models.CashAdvanceUsage) error {
	return e.db.Model(&models.CashAdvanceUsage{}).
		Where("id = ?", id).
		Updates(input).Error
}

// DeleteCashAdvanceUsage deletes an existing cash advance usage record in the database.
//
// The function takes the ID of the record to be deleted as input and attempts to delete
// the corresponding record in the database. It returns an error if the deletion fails,
// otherwise returns nil.
func (e *EmployeeCashAdvanceService) DeleteCashAdvanceUsage(id string) error {
	return e.db.Where("id = ?", id).Delete(&models.CashAdvanceUsage{}).Error
}

// CreateCashAdvanceRefund adds a new cash advance refund record to the database.
//
// The function takes a CashAdvanceRefund model as input and attempts to create
// a new record in the database. It returns an error if the creation fails,
// otherwise returns nil.
func (e *EmployeeCashAdvanceService) CreateCashAdvanceRefund(cashAdvanceRefund *models.CashAdvanceRefund) error {
	return e.db.Create(cashAdvanceRefund).Error
}

// DeleteCashAdvanceRefund deletes an existing cash advance refund record in the database.
//
// The function takes the ID of the record to be deleted as input and attempts to delete
// the corresponding record in the database. It returns an error if the deletion fails,
// otherwise returns nil.
func (e *EmployeeCashAdvanceService) DeleteCashAdvanceRefund(id string) error {
	return e.db.Where("id = ?", id).Delete(&models.CashAdvanceRefund{}).Error
}
