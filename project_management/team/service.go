package team

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

// TeamService provides methods for creating, updating, deleting, and retrieving teams.
//
// The service requires a Gorm database instance and an ERP context.
type TeamService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

// NewTeamService creates a new team service using the provided context.
func NewTeamService(ctx *context.ERPContext) *TeamService {
	return &TeamService{
		db:  ctx.DB,
		ctx: ctx,
	}
}

// CreateTeam creates a new team in the database.
//
// It takes a pointer to a TeamModel as a parameter and returns an error. The function
// uses the gorm.DB connection to create a new record in the teams table. If the
// operation fails, an error is returned.
func (s *TeamService) CreateTeam(data *models.TeamModel) error {
	return s.db.Create(data).Error
}

// UpdateTeam updates a team in the database.
//
// It takes an ID and a pointer to a TeamModel as parameters and returns an error.
// The function uses GORM to update a record in the teams table where the ID matches.
// If the update is successful, the error is nil. Otherwise, the error contains
// information about what went wrong.
func (s *TeamService) UpdateTeam(id string, data *models.TeamModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

// DeleteTeam deletes a team from the database by ID.
//
// It takes an ID as input and returns an error if the deletion operation fails.
// The function uses GORM to delete the team data from the teams table. If the
// deletion is successful, the error is nil. Otherwise, the error contains
// information about what went wrong.
func (s *TeamService) DeleteTeam(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.TeamModel{}).Error
}

// GetTeamByID retrieves a team by its ID from the database.
//
// The team is queried using the GORM First method, and any errors are
// returned to the caller. If the team is not found, a nil pointer is returned
// together with a gorm.ErrRecordNotFound error.
func (s *TeamService) GetTeamByID(id string) (*models.TeamModel, error) {
	var invoice models.TeamModel
	err := s.db.Where("id = ?", id).First(&invoice).Error
	return &invoice, err
}

// GetTeams retrieves a paginated list of teams.
//
// The method uses GORM to query the database for teams, preloading the associated
// Company model. It applies a filter based on the company ID provided in the HTTP
// request header, and another filter based on the search parameter if provided.
// The function utilizes pagination to manage the result set and applies any
// necessary request modifications using the utils.FixRequest utility.
//
// The function returns a paginated page of TeamModel and an error if the
// operation fails.
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
