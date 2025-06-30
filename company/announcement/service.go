package announcement

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type AnnouncementService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewAnnouncementService(db *gorm.DB, ctx *context.ERPContext) *AnnouncementService {
	return &AnnouncementService{db: db, ctx: ctx}
}

func (s *AnnouncementService) CreateAnnouncement(data *models.AnnouncementModel) error {
	return s.db.Create(data).Error
}

func (s *AnnouncementService) UpdateAnnouncement(id string, data *models.AnnouncementModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

func (s *AnnouncementService) DeleteAnnouncement(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.AnnouncementModel{}).Error
}

func (s *AnnouncementService) GetAnnouncementByID(id string) (*models.AnnouncementModel, error) {
	var branch models.AnnouncementModel
	if err := s.db.Where("id = ?", id).First(&branch).Error; err != nil {
		return nil, err
	}
	return &branch, nil
}

func (s *AnnouncementService) FindAllAnnouncementsByEmployee(request *http.Request, employee *models.EmployeeModel) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.Model(&models.AnnouncementModel{})
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}

	if employee != nil {
		stmt = stmt.Joins("LEFT JOIN announcement_employees ON announcement_employees.announcement_model_id = announcements.id")
		stmt = stmt.Where("announcement_employees.employee_model_id = ? or announcement_employees.employee_model_id IS NULL", employee.ID)

		if employee.JobTitleID != nil {
			stmt = stmt.Where("job_title_id = ? or job_title_id IS NULL", employee.JobTitleID)
		}
		if employee.OrganizationID != nil {
			stmt = stmt.Where("organization_id = ? or organization_id IS NULL", employee.OrganizationID)
		}

		if employee.BranchID != nil {
			stmt = stmt.Where("branch_id = ? or branch_id IS NULL", employee.BranchID)
		}
		if employee.WorkLocationID != nil {
			stmt = stmt.Where("work_location_id = ? or work_location_id IS NULL", employee.WorkLocationID)
		}
		if employee.WorkShiftID != nil {
			stmt = stmt.Where("work_shift_id = ? or work_shift_id IS NULL", employee.WorkShiftID)
		}
		stmt = stmt.Group("announcements.id")

	}
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.AnnouncementModel{})
	page.Page = page.Page + 1
	return page, nil
}

func (s *AnnouncementService) FindAllAnnouncements(request *http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.Model(&models.AnnouncementModel{})
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.AnnouncementModel{})
	page.Page = page.Page + 1
	return page, nil
}
