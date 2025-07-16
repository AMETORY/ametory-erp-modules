package citizen

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
)

// CitizenService is a service for interacting with citizens.
//
// It provides methods for creating, retrieving, updating and deleting citizens.
//
// Citizens are a collection of people who are served by the permit hub.
type CitizenService struct {
	ctx *context.ERPContext
}

// NewCitizenService creates a new instance of CitizenService with the given database connection and context.
func NewCitizenService(ctx *context.ERPContext) *CitizenService {
	return &CitizenService{
		ctx: ctx,
	}

}

// GetAllCitizens returns a list of all citizens.
//
// This method uses the paginate package to create a paginated result set.
// It takes a request object which contains the query parameters for the request.
// The Response method is used to populate the result set.
func (s *CitizenService) GetAllCitizens(request *http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.ctx.DB.Model(&models.Citizen{})
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.Citizen{})
	page.Page = page.Page + 1
	return page, nil
}

// CreateCitizen creates a new citizen.
//
// This method takes a citizen model as a parameter and creates a new citizen in the database.
func (s *CitizenService) CreateCitizen(citizen *models.Citizen) error {
	if err := s.ctx.DB.Create(citizen).Error; err != nil {
		return err
	}
	return nil
}

// GetCitizenByID returns a citizen by its id.
//
// This method takes the id of the citizen as a parameter and returns the corresponding citizen model.
func (s *CitizenService) GetCitizenByID(id string) (*models.Citizen, error) {
	var citizen models.Citizen
	if err := s.ctx.DB.Where("id = ?", id).First(&citizen).Error; err != nil {
		return nil, err
	}
	return &citizen, nil
}

// GetCitizenByNIK returns a citizen by its NIK.
//
// This method takes the NIK of the citizen as a parameter and returns the corresponding citizen model.
func (s *CitizenService) GetCitizenByNIK(nik string) (*models.Citizen, error) {
	var citizen models.Citizen
	if err := s.ctx.DB.Where("nik = ?", nik).First(&citizen).Error; err != nil {
		return nil, err
	}
	return &citizen, nil
}

// UpdateCitizen updates a citizen.
//
// This method takes the id of the citizen and a citizen model as parameters and updates the corresponding citizen in the database.
func (s *CitizenService) UpdateCitizen(id string, citizen *models.Citizen) error {
	return s.ctx.DB.Model(&models.Citizen{}).Where("id = ?", id).Updates(citizen).Error
}

// DeleteCitizen deletes a citizen.
//
// This method takes the id of the citizen as a parameter and deletes the corresponding citizen in the database.
func (s *CitizenService) DeleteCitizen(id string) error {
	return s.ctx.DB.Where("id = ?", id).Delete(&models.Citizen{}).Error
}
