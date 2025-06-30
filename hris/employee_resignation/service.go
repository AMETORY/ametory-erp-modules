package employee_resignation

import (
	"net/http"
	"time"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type EmployeeResignationService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewEmployeeResignationService(ctx *context.ERPContext) *EmployeeResignationService {
	return &EmployeeResignationService{db: ctx.DB, ctx: ctx}
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.EmployeeResignation{},
	)
}

func (e *EmployeeResignationService) CreateEmployeeResignation(employeeResignation *models.EmployeeResignation) error {
	return e.db.Create(employeeResignation).Error
}

func (e *EmployeeResignationService) GetEmployeeResignationByID(id string) (*models.EmployeeResignation, error) {
	var employeeResignation models.EmployeeResignation
	err := e.db.
		Preload("Employee").
		Preload("Company").
		Preload("Approver.User").
		Where("id = ?", id).First(&employeeResignation).Error
	if err != nil {
		return nil, err
	}

	return &employeeResignation, nil
}

func (e *EmployeeResignationService) UpdateEmployeeResignation(employeeResignation *models.EmployeeResignation) error {
	return e.db.Save(employeeResignation).Error
}

func (e *EmployeeResignationService) DeleteEmployeeResignation(id string) error {
	return e.db.Delete(&models.EmployeeResignation{}, "id = ?", id).Error
}

func (e *EmployeeResignationService) FindAllEmployeeResignations(request *http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := e.db.
		Preload("Employee").
		Preload("Company").
		Preload("Approver.User").
		Model(&models.EmployeeResignation{})
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	if request.Header.Get("ID-User") != "" {
		stmt = stmt.Where("user_id = ?", request.Header.Get("ID-User"))
	}

	if request.URL.Query().Get("search") != "" {
		stmt = stmt.Where("reason LIKE ?", "%"+request.URL.Query().Get("search")+"%")
	}
	if request.URL.Query().Get("start_date") != "" && request.URL.Query().Get("end_date") != "" {
		stmt = stmt.Where("resignation_date >= ? AND resignation_date <= ?", request.URL.Query().Get("start_date"), request.URL.Query().Get("end_date"))
	}
	if request.URL.Query().Get("date") != "" {
		stmt = stmt.Where("DATE(resignation_date) = ?", request.URL.Query().Get("date"))
	}
	if request.URL.Query().Get("approver_id") != "" {
		stmt = stmt.Where("approver_id = ?", request.URL.Query().Get("approver_id"))
	}

	if request.URL.Query().Get("order") != "" {
		stmt = stmt.Order(request.URL.Query().Get("order"))
	} else {
		stmt = stmt.Order("resignation_date DESC")
	}
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.EmployeeResignation{})
	page.Page = page.Page + 1
	return page, nil
}

func (e *EmployeeResignationService) CountByEmployeeID(employeeID string, startDate *time.Time, endDate *time.Time) (map[string]int64, error) {
	var countREQUESTED, countAPPROVED, countREJECTED int64
	counts := make(map[string]int64)
	e.db.Model(&models.EmployeeResignation{}).
		Where("employee_id = ? AND status = ? AND date >= ? AND date <= ?", employeeID, "REQUESTED", startDate, endDate).
		Count(&countREQUESTED)
	e.db.Model(&models.EmployeeResignation{}).
		Where("employee_id = ? AND status = ? AND date >= ? AND date <= ?", employeeID, "APPROVED", startDate, endDate).
		Count(&countAPPROVED)
	e.db.Model(&models.EmployeeResignation{}).
		Where("employee_id = ? AND status = ? AND date >= ? AND date <= ?", employeeID, "REJECTED", startDate, endDate).
		Count(&countREJECTED)

	counts["REQUESTED"] = countREQUESTED
	counts["APPROVED"] = countAPPROVED
	counts["REJECTED"] = countREJECTED

	return counts, nil
}
