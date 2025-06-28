package work_location

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type WorkLocationService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewWorkLocationService(db *gorm.DB, ctx *context.ERPContext) *WorkLocationService {
	return &WorkLocationService{db: db, ctx: ctx}
}

func (s *WorkLocationService) CreateWorkLocation(data *models.WorkLocationModel) error {
	return s.db.Create(data).Error
}

func (s *WorkLocationService) UpdateWorkLocation(id string, data *models.WorkLocationModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

func (s *WorkLocationService) DeleteWorkLocation(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.WorkLocationModel{}).Error
}

func (s *WorkLocationService) GetWorkLocationByID(id string) (*models.WorkLocationModel, error) {
	var branch models.WorkLocationModel
	if err := s.db.Where("id = ?", id).First(&branch).Error; err != nil {
		return nil, err
	}
	return &branch, nil
}

func (s *WorkLocationService) FindAllWorkLocations(request *http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.Model(&models.WorkLocationModel{})
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.WorkLocationModel{})
	page.Page = page.Page + 1
	return page, nil
}

func (s *WorkLocationService) GetWorkLocationByEmployee(employee *models.EmployeeModel) (*models.WorkLocationModel, error) {
	if employee == nil {
		return nil, nil
	}

	if err := s.db.Model(&employee).Preload("WorkLocation").Find(employee).Error; err != nil {
		return nil, err
	}
	return employee.WorkLocation, nil
}
