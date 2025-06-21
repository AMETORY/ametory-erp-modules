package organization

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
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
	return s.db.Where("id = ?", id).Delete(&models.OrganizationModel{}).Error
}

func (s *OrganizationService) GetOrganizationByID(id string) (*models.OrganizationModel, error) {
	var branch models.OrganizationModel
	if err := s.db.Where("id = ?", id).First(&branch).Error; err != nil {
		return nil, err
	}
	return &branch, nil
}

func (s *OrganizationService) FindAllOrganizations(request *http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.Model(&models.OrganizationModel{})
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.OrganizationModel{})
	page.Page = page.Page + 1
	return page, nil
}
