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

func (s *OrganizationService) CreateOrganization(data *models.OrganizationModel) error {
	return s.db.Create(data).Error
}

func (s *OrganizationService) UpdateOrganization(id string, data *models.OrganizationModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

func (s *OrganizationService) DeleteOrganization(id string) error {
	return s.db.Where("id = ?", id).Unscoped().Delete(&models.OrganizationModel{}).Error
}

func (s *OrganizationService) GetOrganizationByID(id string) (*models.OrganizationModel, error) {
	var branch models.OrganizationModel
	if err := s.db.Where("id = ?", id).First(&branch).Error; err != nil {
		return nil, err
	}
	return &branch, nil
}

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
