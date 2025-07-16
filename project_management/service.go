package project_management

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/project_management/member"
	"github.com/AMETORY/ametory-erp-modules/project_management/project"
	"github.com/AMETORY/ametory-erp-modules/project_management/task"
	"github.com/AMETORY/ametory-erp-modules/project_management/task_attribute"
	"github.com/AMETORY/ametory-erp-modules/project_management/team"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
)

type ProjectManagementService struct {
	ctx                  *context.ERPContext
	ProjectService       *project.ProjectService
	TeamService          *team.TeamService
	MemberService        *member.MemberService
	TaskService          *task.TaskService
	TaskAttributeService *task_attribute.TaskAttributeService
}

// NewProjectManagementService creates a new instance of ProjectManagementService.
//
// It takes an ERPContext as parameter and returns a pointer to a ProjectManagementService.
//
// It uses the ERPContext to initialize the ProjectService, TeamService, MemberService, TaskService, and TaskAttributeService.
// Additionally, it calls the Migrate method of the ProjectManagementService to create the necessary database schema,
// unless SkipMigration is set to true in the ERPContext.
func NewProjectManagementService(ctx *context.ERPContext) *ProjectManagementService {
	service := ProjectManagementService{
		ctx:                  ctx,
		ProjectService:       project.NewProjectService(ctx),
		TeamService:          team.NewTeamService(ctx),
		MemberService:        member.NewMemberService(ctx),
		TaskService:          task.NewTaskService(ctx),
		TaskAttributeService: task_attribute.NewTaskAttibuteService(ctx),
	}
	if !ctx.SkipMigration {
		service.Migrate()
	}
	return &service
}

// Migrate migrates the database schema for the ProjectManagementService.
//
// If the SkipMigration flag is set to true in the context, this method
// will not perform any migration and will return nil. Otherwise, it will
// attempt to auto-migrate the database to include the ProjectModel,
// ProjectActivityModel, ColumnModel, ColumnAction, TeamModel, MemberModel,
// MemberInvitationModel, TaskModel, TaskCommentModel, and TaskAttributeModel
// schemas. If the migration process encounters an error, it will return that
// error. Otherwise, it will return nil upon successful migration.
func (cs *ProjectManagementService) Migrate() error {
	if cs.ctx.SkipMigration {
		return nil
	}
	return cs.ctx.DB.AutoMigrate(
		&models.ProjectModel{},
		&models.ProjectActivityModel{},
		&models.ColumnModel{},
		&models.ColumnAction{},
		&models.TeamModel{},
		&models.MemberModel{},
		&models.MemberInvitationModel{},
		&models.TaskModel{},
		&models.TaskCommentModel{},
		&models.TaskAttributeModel{},
	)
}
