package company

import (
	"fmt"
	"log"
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/company/branch"
	"github.com/AMETORY/ametory-erp-modules/company/organization"
	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type CompanyService struct {
	ctx                 *context.ERPContext
	BranchService       *branch.BranchService
	OrganizationService *organization.OrganizationService
}

func NewCompanyService(ctx *context.ERPContext) *CompanyService {
	fmt.Println("INIT COMPANY SERVICE")
	var service = CompanyService{ctx: ctx,
		OrganizationService: organization.NewOrganizationService(ctx.DB, ctx),
		BranchService:       branch.NewBranchService(ctx.DB, ctx),
	}
	err := service.Migrate()
	if err != nil {
		log.Println("ERROR COMPANY MIGRATE", err)
		panic(err)
	}
	return &service
}

func (s *CompanyService) Migrate() error {
	if s.ctx.SkipMigration {
		return nil
	}
	// s.ctx.DB.Migrator().AlterColumn(&models.CompanyModel{}, "status")
	return s.ctx.DB.AutoMigrate(
		&models.CompanyModel{},
		&models.CompanySector{},
		&models.CompanyCategory{},
		&models.BranchModel{},
		&models.OrganizationModel{},
	)
}

func (s *CompanyService) DB() *gorm.DB {
	return s.ctx.DB
}

func (s *CompanyService) CreateCompany(data *models.CompanyModel) error {
	return s.ctx.DB.Create(data).Error
}

func (s *CompanyService) UpdateCompany(id string, data *models.CompanyModel) error {
	return s.ctx.DB.Where("id = ?", id).Updates(data).Error
}

func (s *CompanyService) DeleteCompany(id string) error {
	return s.ctx.DB.Where("id = ?", id).Delete(&models.CompanyModel{}).Error
}

func (s *CompanyService) GetCompanyByID(id string) (*models.CompanyModel, error) {
	var company models.CompanyModel
	err := s.ctx.DB.Where("id = ?", id).First(&company).Error
	return &company, err
}

func (s *CompanyService) GetCompanyUsers(id string, request http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.ctx.DB.Model(&models.UserModel{}).Preload("Roles", "roles.company_id = ?", id).
		Joins("JOIN user_companies ON user_companies.user_model_id = users.id").
		Where("user_companies.company_model_id = ?", id)
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(&request).Response(&[]models.UserModel{})
	return page, nil
}
func (s *CompanyService) GetCompanyWithMerchantsByID(id string) (*models.CompanyModel, error) {
	var company models.CompanyModel
	err := s.ctx.DB.Where("id = ?", id).First(&company).Error
	if err != nil {
		return nil, err
	}
	var merchants []models.MerchantModel
	var warehouses []models.WarehouseModel
	err = s.ctx.DB.Where("company_id = ?", id).Find(&merchants).Error
	if err != nil {
		return nil, err
	}
	err = s.ctx.DB.Where("company_id = ?", id).Find(&warehouses).Error
	if err != nil {
		return nil, err
	}
	company.Merchants = merchants
	company.Warehouses = warehouses
	return &company, nil
}

func (s *CompanyService) GetCompanyByCode(code string) (*models.CompanyModel, error) {
	var company models.CompanyModel
	err := s.ctx.DB.Where("code = ?", code).First(&company).Error
	return &company, err
}

func (s *CompanyService) GetCompanies(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.ctx.DB
	if search != "" {
		stmt = stmt.Where("companies.address ILIKE ? OR companies.name ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
		)
	}
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	request.URL.Query().Get("page")
	stmt = stmt.Model(&models.CompanyModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.CompanyModel{})
	page.Page = page.Page + 1
	return page, nil
}

func (s *CompanyService) GetCategories(sectorID *string) ([]models.CompanyCategory, error) {
	var categories []models.CompanyCategory
	db := s.ctx.DB
	if sectorID != nil {
		db = db.Where("sector_id = ?", sectorID)
	}
	err := db.Find(&categories).Error
	return categories, err
}

func (s *CompanyService) GetSectors() ([]models.CompanySector, error) {
	var sectors []models.CompanySector
	err := s.ctx.DB.Preload("Categories").Find(&sectors).Error
	return sectors, err
}

func (s *CompanyService) GetCompaniesByUserID(userID string) ([]models.CompanyModel, error) {
	var companies []models.CompanyModel
	err := s.ctx.DB.Joins("JOIN user_companies AS uc ON companies.id = uc.company_model_id").Where("uc.user_model_id = ?", userID).Find(&companies).Error
	return companies, err
}
