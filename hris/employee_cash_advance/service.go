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

func (e *EmployeeCashAdvanceService) UpdateEmployeeCashAdvance(employeeCashAdvance *models.EmployeeCashAdvance) error {
	return e.db.Save(employeeCashAdvance).Error
}

func (e *EmployeeCashAdvanceService) DeleteEmployeeCashAdvance(id string) error {
	return e.db.Delete(&models.EmployeeCashAdvance{}, "id = ?", id).Error
}

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

func (e *EmployeeCashAdvanceService) CreateCashAdvanceUsage(cashAdvanceUsage *models.CashAdvanceUsage) error {
	return e.db.Create(cashAdvanceUsage).Error
}

func (e *EmployeeCashAdvanceService) UpdateEmployeeCashAdvanceUsage(id string, input *models.CashAdvanceUsage) error {
	return e.db.Model(&models.CashAdvanceUsage{}).
		Where("id = ?", id).
		Updates(input).Error
}

func (e *EmployeeCashAdvanceService) DeleteCashAdvanceUsage(id string) error {
	return e.db.Where("id = ?", id).Delete(&models.CashAdvanceUsage{}).Error
}

func (e *EmployeeCashAdvanceService) CreateCashAdvanceRefund(cashAdvanceRefund *models.CashAdvanceRefund) error {
	return e.db.Create(cashAdvanceRefund).Error
}

func (e *EmployeeCashAdvanceService) DeleteCashAdvanceRefund(id string) error {
	return e.db.Where("id = ?", id).Delete(&models.CashAdvanceRefund{}).Error
}
