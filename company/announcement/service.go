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

// NewAnnouncementService returns a new instance of AnnouncementService.
//
// The service is created by providing a GORM database instance and an ERP context.
// The ERP context is used for authentication and authorization purposes, while the
// database instance is used for CRUD (Create, Read, Update, Delete) operations.
func NewAnnouncementService(db *gorm.DB, ctx *context.ERPContext) *AnnouncementService {
	return &AnnouncementService{db: db, ctx: ctx}
}

// CreateAnnouncement creates a new announcement record in the database.
//
// It takes a pointer to an AnnouncementModel as input and returns an error if
// the operation fails. The function uses GORM to insert the announcement data
// into the announcements table.

func (s *AnnouncementService) CreateAnnouncement(data *models.AnnouncementModel) error {
	return s.db.Create(data).Error
}

// UpdateAnnouncement updates an existing announcement record in the database.
//
// It takes an ID and a pointer to an AnnouncementModel as input and returns an
// error if the operation fails. The function uses GORM to update the existing
// announcement data in the announcements table.
func (s *AnnouncementService) UpdateAnnouncement(id string, data *models.AnnouncementModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

// DeleteAnnouncement deletes an existing announcement record from the database.
//
// It takes an ID as input and returns an error if the operation fails. The
// function uses GORM to delete the existing announcement data from the
// announcements table.
func (s *AnnouncementService) DeleteAnnouncement(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.AnnouncementModel{}).Error
}

// GetAnnouncementByID retrieves an announcement from the database by ID.
//
// It takes an ID as input and returns an AnnouncementModel and an error. The
// function uses GORM to retrieve the announcement data from the announcements
// table. If the operation fails, an error is returned.
func (s *AnnouncementService) GetAnnouncementByID(id string) (*models.AnnouncementModel, error) {
	var branch models.AnnouncementModel
	if err := s.db.Where("id = ?", id).First(&branch).Error; err != nil {
		return nil, err
	}
	return &branch, nil
}

// FindAllAnnouncementsByEmployee retrieves a paginated list of announcements
// associated with a specific employee.
//
// It takes an HTTP request and an EmployeeModel as input and returns a paginated
// Page of AnnouncementModel and an error if the operation fails. The function
// applies various filters based on the employee's attributes such as job title,
// organization, branch, work location, and work shift. It also checks for the
// company ID in the request header to further filter the announcements. The
// results are grouped by announcement ID to avoid duplicates.

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

// FindAllAnnouncements retrieves a paginated list of announcements associated with
// a specific company.
//
// It takes an HTTP request as input and returns a paginated Page of
// AnnouncementModel and an error if the operation fails. The function applies a
// filter based on the company ID in the request header to further filter the
// announcements.
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
