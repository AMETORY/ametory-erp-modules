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

func NewJobTitleService(ctx *context.ERPContext) *JobTitleService {
	return &JobTitleService{db: ctx.DB, ctx: ctx}
}

func (e *JobTitleService) CreateJobTitle(jobTitle *models.JobTitleModel) error {
	return e.db.Create(jobTitle).Error
}

func (e *JobTitleService) GetJobTitleByID(id string) (*models.JobTitleModel, error) {
	var jobTitle models.JobTitleModel
	err := e.db.First(&jobTitle, id).Error
	if err != nil {
		return nil, err
	}
	return &jobTitle, nil
}

func (e *JobTitleService) UpdateJobTitle(jobTitle *models.JobTitleModel) error {
	return e.db.Save(jobTitle).Error
}

func (e *JobTitleService) DeleteJobTitle(id string) error {
	return e.db.Delete(&models.JobTitleModel{}, id).Error
}

func (e *JobTitleService) FindAllJobTitles(request *http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := e.db.Model(&models.JobTitleModel{})
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.JobTitleModel{})
	page.Page = page.Page + 1
	return page, nil
}
