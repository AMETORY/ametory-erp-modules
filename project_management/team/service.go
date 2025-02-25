package team

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type TeamService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewTeamService(ctx *context.ERPContext) *TeamService {
	return &TeamService{
		db:  ctx.DB,
		ctx: ctx,
	}
}

func (s *TeamService) CreateTeam(data *models.TeamModel) error {
	return s.db.Create(data).Error
}

func (s *TeamService) UpdateTeam(id string, data *models.TeamModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

func (s *TeamService) DeleteTeam(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.TeamModel{}).Error
}

func (s *TeamService) GetTeamByID(id string) (*models.TeamModel, error) {
	var invoice models.TeamModel
	err := s.db.Where("id = ?", id).First(&invoice).Error
	return &invoice, err
}

func (s *TeamService) GetTeams(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db
	if search != "" {
		stmt = stmt.Where("teams.name ILIKE ?",
			"%"+search+"%",
		)
	}
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	request.URL.Query().Get("page")
	stmt = stmt.Model(&models.TeamModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.TeamModel{})
	page.Page = page.Page + 1
	return page, nil
}
