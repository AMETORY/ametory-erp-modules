package project_management

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/project_management/member"
	"github.com/AMETORY/ametory-erp-modules/project_management/project"
	"github.com/AMETORY/ametory-erp-modules/project_management/team"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
)

type ProjectManagementService struct {
	ctx            *context.ERPContext
	ProjectService *project.ProjectService
	TeamService    *team.TeamService
	MemberService  *member.MemberService
}

func NewProjectManagementService(ctx *context.ERPContext) *ProjectManagementService {
	service := ProjectManagementService{
		ctx:            ctx,
		ProjectService: project.NewProjectService(ctx),
		TeamService:    team.NewTeamService(ctx),
		MemberService:  member.NewMemberService(ctx),
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
		&models.ColumnModel{},
		&models.TeamModel{},
		&models.MemberModel{},
		&models.MemberInvitationModel{},
		&models.TaskModel{},
		&models.TaskCommentModel{},
	)
}
