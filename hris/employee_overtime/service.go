package employee_overtime

import (
	"net/http"
	"time"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type EmployeeOvertimeService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewEmployeeOvertimeService(ctx *context.ERPContext) *EmployeeOvertimeService {
	return &EmployeeOvertimeService{db: ctx.DB, ctx: ctx}
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.EmployeeOvertimeModel{},
	)
}

func (e *EmployeeOvertimeService) CreateEmployeeOvertime(employeeOvertime *models.EmployeeOvertimeModel) error {
	return e.db.Create(employeeOvertime).Error
}

func (e *EmployeeOvertimeService) GetEmployeeOvertimeByID(id string) (*models.EmployeeOvertimeModel, error) {
	var employeeOvertime models.EmployeeOvertimeModel
	err := e.db.
		Preload("Employee", func(tx *gorm.DB) *gorm.DB {
			return tx.Preload("User").
				Preload("JobTitle").
				Preload("WorkLocation").
				Preload("WorkShift").
				Preload("Branch")
		}).
		Preload("Approver.User").
		Preload("Attendance").
		Preload("ApprovalByAdmin").
		Where("id = ?", id).First(&employeeOvertime).Error
	if err != nil {
		return nil, err
	}
	return &employeeOvertime, nil
}

func (e *EmployeeOvertimeService) UpdateEmployeeOvertime(employeeOvertime *models.EmployeeOvertimeModel) error {
	return e.db.Save(employeeOvertime).Error
}

func (e *EmployeeOvertimeService) DeleteEmployeeOvertime(id string) error {
	return e.db.Delete(&models.EmployeeOvertimeModel{}, "id = ?", id).Error
}

func (e *EmployeeOvertimeService) FindAllByEmployeeID(request *http.Request, employeeID string) (paginate.Page, error) {
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
		Where("employee_id = ?", employeeID).
		Model(&models.EmployeeOvertimeModel{})
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}

	if request.URL.Query().Get("search") != "" {
		stmt = stmt.Where("reason LIKE ?", "%"+request.URL.Query().Get("search")+"%")
	}
	if request.URL.Query().Get("start_date") != "" && request.URL.Query().Get("end_date") != "" {
		stmt = stmt.Where("start_time_request >= ? AND end_time_request <= ?", request.URL.Query().Get("start_date"), request.URL.Query().Get("end_date"))
	} else if request.URL.Query().Get("start_date") != "" {
		stmt = stmt.Where("start_time_request = ?", request.URL.Query().Get("start_date"))
	}
	if request.URL.Query().Get("date") != "" {
		stmt = stmt.Where("DATE(start_time_request) = ?", request.URL.Query().Get("date"))
	}
	if request.URL.Query().Get("approver_id") != "" {
		stmt = stmt.Where("approver_id = ?", request.URL.Query().Get("approver_id"))
	}
	if request.URL.Query().Get("reviewer_id") != "" {
		stmt = stmt.Where("reviewer_id = ?", request.URL.Query().Get("reviewer_id"))
	}
	if request.URL.Query().Get("order") != "" {
		stmt = stmt.Order(request.URL.Query().Get("order"))
	} else {
		stmt = stmt.Order("start_time_request DESC")
	}
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.EmployeeOvertimeModel{})
	page.Page = page.Page + 1
	return page, nil
}
func (e *EmployeeOvertimeService) FindAllEmployeeOvertimes(request *http.Request) (paginate.Page, error) {
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
		Model(&models.EmployeeOvertimeModel{})
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}

	if request.URL.Query().Get("search") != "" {
		stmt = stmt.Where("reason LIKE ?", "%"+request.URL.Query().Get("search")+"%")
	}
	if request.URL.Query().Get("start_date") != "" && request.URL.Query().Get("end_date") != "" {
		stmt = stmt.Where("start_time_request >= ? AND end_time_request <= ?", request.URL.Query().Get("start_date"), request.URL.Query().Get("end_date"))
	} else if request.URL.Query().Get("start_date") != "" {
		stmt = stmt.Where("start_time_request = ?", request.URL.Query().Get("start_date"))
	}
	if request.URL.Query().Get("date") != "" {
		stmt = stmt.Where("DATE(start_time_request) = ?", request.URL.Query().Get("date"))
	}
	if request.URL.Query().Get("approver_id") != "" {
		stmt = stmt.Where("approver_id = ?", request.URL.Query().Get("approver_id"))
	}
	if request.URL.Query().Get("reviewer_id") != "" {
		stmt = stmt.Where("reviewer_id = ?", request.URL.Query().Get("reviewer_id"))
	}
	if request.URL.Query().Get("order") != "" {
		stmt = stmt.Order(request.URL.Query().Get("order"))
	} else {
		stmt = stmt.Order("start_time_request DESC")
	}
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.EmployeeOvertimeModel{})
	page.Page = page.Page + 1
	return page, nil
}

func (e *EmployeeOvertimeService) CountByEmployeeID(employeeID string, startDate *time.Time, endDate *time.Time) (map[string]int64, error) {
	var countPENDING, countAPPROVED, countREJECTED int64
	counts := make(map[string]int64)
	e.db.Model(&models.EmployeeOvertimeModel{}).
		Where("employee_id = ? AND status = ? AND start_time_request >= ? AND end_time_request <= ?", employeeID, "PENDING", startDate, endDate).
		Count(&countPENDING)
	e.db.Model(&models.EmployeeOvertimeModel{}).
		Where("employee_id = ? AND status = ? AND start_time_request >= ? AND end_time_request <= ?", employeeID, "APPROVED", startDate, endDate).
		Count(&countAPPROVED)
	e.db.Model(&models.EmployeeOvertimeModel{}).
		Where("employee_id = ? AND status = ? AND start_time_request >= ? AND end_time_request <= ?", employeeID, "REJECTED", startDate, endDate).
		Count(&countREJECTED)

	counts["PENDING"] = countPENDING
	counts["APPROVED"] = countAPPROVED
	counts["REJECTED"] = countREJECTED

	return counts, nil
}
