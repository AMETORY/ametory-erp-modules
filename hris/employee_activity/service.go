package employee_activity

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type EmployeeActivityService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewEmployeeActivityService(ctx *context.ERPContext) *EmployeeActivityService {
	return &EmployeeActivityService{db: ctx.DB, ctx: ctx}
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.EmployeeActivityModel{},
	)
}

func (service *EmployeeActivityService) CreateEmployeeActivity(activity *models.EmployeeActivityModel) error {
	// utils.LogJson(activity.AssignedEmployees)
	return service.db.Create(activity).Error
}

func (service *EmployeeActivityService) GetEmployeeActivityByID(id string) (*models.EmployeeActivityModel, error) {
	var activity models.EmployeeActivityModel
	err := service.db.First(&activity, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &activity, nil
}

func (service *EmployeeActivityService) UpdateEmployeeActivity(activity *models.EmployeeActivityModel) error {
	return service.db.Save(activity).Error
}

func (service *EmployeeActivityService) DeleteEmployeeActivity(id string) error {
	return service.db.Delete(&models.EmployeeActivityModel{}, "id = ?", id).Error
}

func (service *EmployeeActivityService) FindAll(request *http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := service.db.Model(&models.EmployeeActivityModel{})
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.EmployeeActivityModel{})
	page.Page = page.Page + 1
	return page, nil
}

func (service *EmployeeActivityService) FindAllByEmployeeID(request *http.Request, employeeID string, activityType string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := service.db.Model(&models.EmployeeActivityModel{}).Where("employee_id = ?", employeeID)
	if activityType != "" {
		stmt = stmt.Where("activity_type = ?", activityType)
	}

	if request.URL.Query().Get("search") != "" {
		stmt = stmt.Where("name LIKE ?", "%"+request.URL.Query().Get("search")+"%")
	}

	if request.URL.Query().Get("start_date") != "" && request.URL.Query().Get("end_date") != "" {
		stmt = stmt.Where("start_date >= ? AND end_date <= ?", request.URL.Query().Get("start_date"), request.URL.Query().Get("end_date"))
	} else if request.URL.Query().Get("start_date") != "" {
		stmt = stmt.Where("start_date = ?", request.URL.Query().Get("start_date"))
	}

	if request.URL.Query().Get("date") != "" {
		stmt = stmt.Where("start_date = ?", request.URL.Query().Get("date"))
	}

	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.EmployeeActivityModel{})
	page.Page = page.Page + 1
	return page, nil
}
