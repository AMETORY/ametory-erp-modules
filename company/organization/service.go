package organization

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"gorm.io/gorm"
)

type OrganizationService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewOrganizationService(db *gorm.DB, ctx *context.ERPContext) *OrganizationService {
	return &OrganizationService{db: db, ctx: ctx}
}

// CreateOrganization adds a new organization record to the database.
// It returns an error if the creation fails.

func (s *OrganizationService) CreateOrganization(data *models.OrganizationModel) error {
	return s.db.Create(data).Error
}

// UpdateOrganization updates an existing organization record in the database.
//
// The function takes an ID and a pointer to a OrganizationModel as input and
// returns an error. The function uses GORM to update the organization data in
// the organizations table. If the update is successful, the error is nil.
// Otherwise, the error contains information about what went wrong.
func (s *OrganizationService) UpdateOrganization(id string, data *models.OrganizationModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

// DeleteOrganization deletes an existing organization from the database.
//
// It takes an ID as input and returns an error. The function uses GORM to delete
// the organization data from the organizations table. If the deletion is
// successful, the error is nil. Otherwise, the error contains information about
// what went wrong.
func (s *OrganizationService) DeleteOrganization(id string) error {
	return s.db.Where("id = ?", id).Unscoped().Delete(&models.OrganizationModel{}).Error
}

// GetOrganizationByID retrieves an organization from the database by ID.
//
// It takes an ID as input and returns an OrganizationModel and an error. The
// function uses GORM to retrieve the organization data from the organizations
// table. If the operation fails, an error is returned.
func (s *OrganizationService) GetOrganizationByID(id string) (*models.OrganizationModel, error) {
	var branch models.OrganizationModel
	if err := s.db.Where("id = ?", id).First(&branch).Error; err != nil {
		return nil, err
	}
	return &branch, nil
}

// FindAllOrganizations retrieves all organizations from the database with
// recursive children.
//
// It takes an HTTP request as input and returns a slice of OrganizationModel and
// an error. The function uses GORM to retrieve all root organizations and then
// recursively find all children until there are no more children. If the
// operation fails, an error is returned.
func (s *OrganizationService) FindAllOrganizations(request *http.Request) ([]models.OrganizationModel, error) {
	orgs := []models.OrganizationModel{}
	stmt := s.db.Model(&models.OrganizationModel{})
	stmt = stmt.Where("parent_id is null")
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	stmt.Find(&orgs)

	// Get all recursive

	for i, org := range orgs {
		children, err := s.getChildren(org.ID)
		if err != nil {
			continue
		}
		org.SubOrganizations = children
		orgs[i] = org
	}
	return orgs, nil
}

// getChildren retrieves all children of a given organization ID recursively.
//
// It takes an ID as input and returns a slice of OrganizationModel and an error.
// The function uses GORM to retrieve all children of the given organization ID
// and then recursively finds all grandchildren until there are no more children.
// If the operation fails, an error is returned.
func (s *OrganizationService) getChildren(id string) ([]models.OrganizationModel, error) {
	var children []models.OrganizationModel
	if err := s.db.Where("parent_id = ?", id).Find(&children).Error; err != nil {
		return nil, err
	}
	for i, child := range children {
		child.SubOrganizations, _ = s.getChildren(child.ID)
		children[i] = child
	}
	return children, nil
}
