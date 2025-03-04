package project

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type ProjectService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewProjectService(ctx *context.ERPContext) *ProjectService {
	return &ProjectService{db: ctx.DB, ctx: ctx}
}

func (s *ProjectService) CreateProject(data *models.ProjectModel) error {
	return s.db.Create(data).Error
}

func (s *ProjectService) UpdateProject(id string, data *models.ProjectModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

func (s *ProjectService) DeleteProject(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.ProjectModel{}).Error
}

func (s *ProjectService) GetProjectByID(id string) (*models.ProjectModel, error) {
	var invoice models.ProjectModel
	err := s.db.Preload("Columns").Preload("Members.User").Where("id = ?", id).First(&invoice).Error
	return &invoice, err
}

func (s *ProjectService) GetProjects(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.Preload("Members.User")
	if search != "" {
		stmt = stmt.Where("projects.description ILIKE ? OR projects.name ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
		)
	}
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	request.URL.Query().Get("page")
	stmt = stmt.Model(&models.ProjectModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.ProjectModel{})
	page.Page = page.Page + 1
	return page, nil
}

func (s *ProjectService) CreateColumn(data *models.ColumnModel) error {
	return s.db.Create(data).Error
}

func (s *ProjectService) UpdateColumn(id string, data *models.ColumnModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

func (s *ProjectService) DeleteColumn(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.ColumnModel{}).Error
}

func (s *ProjectService) GetColumnByID(id string) (*models.ColumnModel, error) {
	var invoice models.ColumnModel
	err := s.db.Where("id = ?", id).First(&invoice).Error
	return &invoice, err
}

func (s *ProjectService) GetColumns(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db
	if search != "" {
		stmt = stmt.Where("columns.name ILIKE ?",
			"%"+search+"%",
		)
	}

	request.URL.Query().Get("page")
	stmt = stmt.Model(&models.ColumnModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.ColumnModel{})
	page.Page = page.Page + 1
	return page, nil
}
