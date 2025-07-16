package employee

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type JobTitleService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

// NewJobTitleService creates a new instance of JobTitleService.
//
// The service is created by providing a GORM database instance and an ERP context.
// The ERP context is used for authentication and authorization purposes, while the
// database instance is used for CRUD (Create, Read, Update, Delete) operations.
func NewJobTitleService(ctx *context.ERPContext) *JobTitleService {
	return &JobTitleService{db: ctx.DB, ctx: ctx}
}

// CreateJobTitle creates a new job title in the database.
//
// The job title is created using the jobTitle parameter, which is a pointer
// to a JobTitleModel struct. The job title is inserted into the database
// using the GORM Create method, and any errors are returned to the caller.
func (e *JobTitleService) CreateJobTitle(jobTitle *models.JobTitleModel) error {
	return e.db.Create(jobTitle).Error
}

// GetJobTitleByID retrieves a job title by its ID from the database.
//
// The job title is queried using the GORM First method, and any errors
// are returned to the caller. If the job title is not found, a nil pointer
// is returned together with a gorm.ErrRecordNotFound error.
func (e *JobTitleService) GetJobTitleByID(id string) (*models.JobTitleModel, error) {
	var jobTitle models.JobTitleModel
	err := e.db.First(&jobTitle, id).Error
	if err != nil {
		return nil, err
	}
	return &jobTitle, nil
}

// UpdateJobTitle updates an existing job title in the database.
//
// The function takes a pointer to a JobTitleModel as input and updates
// the corresponding record in the database. It uses the GORM Save method
// to persist the changes. If the update is successful, the function returns
// nil. Otherwise, it returns an error indicating the reason for failure.
func (e *JobTitleService) UpdateJobTitle(jobTitle *models.JobTitleModel) error {
	return e.db.Save(jobTitle).Error
}

// DeleteJobTitle deletes a job title from the database by its ID.
//
// It takes an ID as input and returns an error if the deletion operation fails.
// The function uses GORM to delete the job title data from the job_titles table.
// If the deletion is successful, the error is nil. Otherwise, the error contains
// information about what went wrong.
func (e *JobTitleService) DeleteJobTitle(id string) error {
	return e.db.Delete(&models.JobTitleModel{}, "id = ?", id).Error
}

// FindAllJobTitles retrieves a paginated list of job titles from the database.
//
// It takes an HTTP request as input and returns a paginated page of JobTitleModel
// and an error if the operation fails. The function applies any necessary request
// modifications using the utils.FixRequest utility and then uses pagination to
// manage the result set. The function returns a paginated page of JobTitleModel and
// an error if the operation fails.
func (e *JobTitleService) FindAllJobTitles(request *http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := e.db.Model(&models.JobTitleModel{})
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.JobTitleModel{})
	page.Page = page.Page + 1
	return page, nil
}
