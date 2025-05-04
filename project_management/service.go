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
