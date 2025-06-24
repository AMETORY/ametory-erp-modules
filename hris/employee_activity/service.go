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

func (service *EmployeeActivityService) FindAllByEmployeeID(request *http.Request, employeeID string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := service.db.Model(&models.EmployeeActivityModel{}).Where("employee_id = ?", employeeID)
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.EmployeeActivityModel{})
	page.Page = page.Page + 1
	return page, nil
}
