package project

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ProjectService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewProjectService(ctx *context.ERPContext) *ProjectService {
	return &ProjectService{db: ctx.DB, ctx: ctx}
}

func (s *ProjectService) CreateProject(data *models.ProjectModel) error {
	return s.db.Omit(clause.Associations).Create(data).Error
}

func (s *ProjectService) UpdateProject(id string, data *models.ProjectModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

func (s *ProjectService) DeleteProject(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.ProjectModel{}).Error
}

func (s *ProjectService) GetProjectByID(id string, memberID *string) (*models.ProjectModel, error) {
	var project models.ProjectModel
	db := s.db.Preload("Columns", func(db *gorm.DB) *gorm.DB {
		return db.Order(`"order" asc`).Preload("Tasks")
	}).Preload("Members.User")
	if memberID != nil {
		db = db.
			Joins("JOIN project_members ON project_members.project_model_id = projects.id").
			// Joins("JOIN members ON members.id = project_members.member_model_id").
			Where("project_members.member_model_id = ?", *memberID)
	}
	err := db.Where("id = ?", id).First(&project).Error
	return &project, err
}

func (s *ProjectService) GetProjects(request http.Request, search string, memberID *string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.Preload("Columns").Preload("Members.User")
	if search != "" {
		stmt = stmt.Where("projects.description ILIKE ? OR projects.name ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
		)
	}
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	if memberID != nil {
		stmt = stmt.
			Joins("JOIN project_members ON project_members.project_model_id = projects.id").
			// Joins("JOIN members ON members.id = project_members.member_model_id").
			Where("project_members.member_model_id = ?", *memberID)
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
	return s.db.Where("id = ?", id).Omit(clause.Associations).Updates(data).Error
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

func (s *ProjectService) AddMemberToProject(projectID string, memberID string) error {
	return s.db.Table("project_members").Create(map[string]interface{}{
		"project_model_id": projectID,
		"member_model_id":  memberID,
	}).Error
}

func (s *ProjectService) GetMembersByProjectID(projectID string) ([]models.MemberModel, error) {
	var project models.ProjectModel
	err := s.db.Model(&models.ProjectModel{}).Where("id = ?", projectID).Preload("Members.User").Find(&project).Error
	return project.Members, err
}

func (s *ProjectService) AddActivity(projectID, memberID string, columnID, taskID *string, activityType string, notes *string) (*models.ProjectActivityModel, error) {
	var activity models.ProjectActivityModel = models.ProjectActivityModel{
		ProjectID:    projectID,
		MemberID:     memberID,
		TaskID:       taskID,
		ColumnID:     columnID,
		ActivityType: activityType,
		Notes:        notes,
	}

	if err := s.db.Create(&activity).Error; err != nil {
		return nil, err
	}

	return &activity, nil
}

func (s *ProjectService) GetRecentActivities(projectID string, limit int) ([]models.ProjectActivityModel, error) {
	var activities []models.ProjectActivityModel
	err := s.db.
		Preload("Project").Preload("Member.User").Preload("Column").Preload("Task").
		Where("project_id = ?", projectID).
		Order("activity_date desc").
		Limit(limit).
		Find(&activities).Error
	return activities, err
}
