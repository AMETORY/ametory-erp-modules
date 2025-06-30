package employee_cash_advance

import (
	"net/http"
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

func NewEmployeeCashAdvanceService(ctx *context.ERPContext) *EmployeeCashAdvanceService {
	return &EmployeeCashAdvanceService{db: ctx.DB, ctx: ctx}
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.EmployeeCashAdvance{},
		&models.CashAdvanceUsage{},
		&models.CashAdvanceRefund{},
	)
}

func (e *EmployeeCashAdvanceService) CreateEmployeeCashAdvance(employeeCashAdvance *models.EmployeeCashAdvance) error {
	return e.db.Create(employeeCashAdvance).Error
}

func (e *EmployeeCashAdvanceService) GetEmployeeCashAdvanceByID(id string) (*models.EmployeeCashAdvance, error) {
	var employeeCashAdvance models.EmployeeCashAdvance
	err := e.db.
		Preload("Employee").
		Preload("Company").
		Preload("Approver.User").
		Where("id = ?", id).First(&employeeCashAdvance).Error
	if err != nil {
		return nil, err
	}
	return &employeeCashAdvance, nil
}

func (e *EmployeeCashAdvanceService) UpdateEmployeeCashAdvance(employeeCashAdvance *models.EmployeeCashAdvance) error {
	return e.db.Save(employeeCashAdvance).Error
}

func (e *EmployeeCashAdvanceService) DeleteEmployeeCashAdvance(id string) error {
	return e.db.Delete(&models.EmployeeCashAdvance{}, "id = ?", id).Error
}

func (e *EmployeeCashAdvanceService) FindAllByEmployeeID(request *http.Request, employeeID string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := e.db.
		Preload("Employee").
		Preload("Company").
		Preload("Approver.User").
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
func (e *EmployeeCashAdvanceService) FindAllEmployeeCashAdvances(request *http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := e.db.
		Preload("Employee").
		Preload("Company").
		Preload("Approver").
		Preload("Reviewer").
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
