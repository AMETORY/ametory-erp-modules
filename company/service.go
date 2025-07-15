package company

import (
	"fmt"
	"log"
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/company/announcement"
	"github.com/AMETORY/ametory-erp-modules/company/branch"
	"github.com/AMETORY/ametory-erp-modules/company/organization"
	"github.com/AMETORY/ametory-erp-modules/company/work_location"
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
	WorkLocationService *work_location.WorkLocationService
	AnnouncementService *announcement.AnnouncementService
}

// NewCompanyService returns a new instance of CompanyService.
//
// The service is created by providing a GORM database instance and an ERP context.
// The ERP context is used for authentication and authorization purposes, while the
// database instance is used for CRUD (Create, Read, Update, Delete) operations.
//
// The service will also create a new instance of BranchService, OrganizationService,
// WorkLocationService, and AnnouncementService.
//
// The service will call the Migrate() method after creation to migrate the database.
// If the migration fails, the service will panic.
func NewCompanyService(ctx *context.ERPContext) *CompanyService {
	fmt.Println("INIT COMPANY SERVICE")
	var service = CompanyService{ctx: ctx,
		OrganizationService: organization.NewOrganizationService(ctx.DB, ctx),
		BranchService:       branch.NewBranchService(ctx.DB, ctx),
		WorkLocationService: work_location.NewWorkLocationService(ctx.DB, ctx),
		AnnouncementService: announcement.NewAnnouncementService(ctx.DB, ctx),
	}
	err := service.Migrate()
	if err != nil {
		log.Println("ERROR COMPANY MIGRATE", err)
		panic(err)
	}
	return &service
}

// Migrate migrates the database for the CompanyService.
//
// If the SkipMigration flag is true, the method will return nil.
// Otherwise, the method will migrate the database by creating the
// tables for CompanyModel, CompanySector, CompanyCategory, BranchModel,
// OrganizationModel, WorkLocationModel, and AnnouncementModel.
//
// If the migration fails, the method will return an error.
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
		&models.WorkLocationModel{},
		&models.AnnouncementModel{},
	)
}

// DB returns the underlying database connection.
//
// The method returns the GORM database connection that is used by the service
// for CRUD (Create, Read, Update, Delete) operations.
func (s *CompanyService) DB() *gorm.DB {
	return s.ctx.DB
}

// CreateCompany creates a new company record in the database.
//
// It takes a pointer to a CompanyModel as input and returns an error if
// the operation fails. The function uses GORM to insert the company data
// into the companies table.

func (s *CompanyService) CreateCompany(data *models.CompanyModel) error {
	return s.ctx.DB.Create(data).Error
}

// UpdateCompany updates an existing company record in the database.
//
// It takes an ID and a pointer to a CompanyModel as input and returns an error
// if the operation fails. The function uses GORM to update the company data
// in the companies table where the ID matches.
//
// If the update is successful, the error is nil. Otherwise, the error contains
// information about what went wrong.
func (s *CompanyService) UpdateCompany(id string, data *models.CompanyModel) error {
	return s.ctx.DB.Where("id = ?", id).Updates(data).Error
}

// DeleteCompany deletes an existing company record from the database.
//
// It takes an ID as input and returns an error if the deletion operation fails.
// The function uses GORM to delete the company data from the companies table.
// If the deletion is successful, the error is nil. Otherwise, the error contains
// information about what went wrong.
func (s *CompanyService) DeleteCompany(id string) error {
	return s.ctx.DB.Where("id = ?", id).Delete(&models.CompanyModel{}).Error
}

// GetCompanyByID retrieves a company record from the database by ID.
//
// It takes an ID as input and returns a pointer to a CompanyModel and an error.
// The function uses GORM to retrieve the company data from the companies table.
// If the operation fails, an error is returned.
func (s *CompanyService) GetCompanyByID(id string) (*models.CompanyModel, error) {
	var company models.CompanyModel
	err := s.ctx.DB.Where("id = ?", id).First(&company).Error
	return &company, err
}

// GetCompanyUsers retrieves a paginated list of users associated with a specific company.
//
// It takes a company ID and an HTTP request as input. The method uses GORM to
// query the database for users linked to the specified company ID. It preloads
// the "Roles" of each user, filtered by the given company ID. The function utilizes
// pagination to manage the result set and applies any necessary request modifications
// using the utils.FixRequest utility.
//
// The function returns a paginated page of UserModel and an error if the operation fails.

func (s *CompanyService) GetCompanyUsers(id string, request http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.ctx.DB.Model(&models.UserModel{}).Preload("Roles", "roles.company_id = ?", id).
		Joins("JOIN user_companies ON user_companies.user_model_id = users.id").
		Where("user_companies.company_model_id = ?", id)
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(&request).Response(&[]models.UserModel{})
	return page, nil
}

// GetCompanyWithMerchantsByID retrieves a company record from the database by ID,
// along with its associated merchants and warehouses.
//
// It takes an ID as input and returns a pointer to a CompanyModel and an error.
// The function uses GORM to retrieve the company data and its associated records
// from the companies, merchants, and warehouses tables.
//
// If the operation fails, an error is returned. Otherwise, the error is nil.
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

// GetCompanyByCode retrieves a company record from the database by code.
//
// It takes a code as input and returns a pointer to a CompanyModel and an error.
// The function uses GORM to retrieve the company data from the companies table.
// If the operation fails, an error is returned. Otherwise, the error is nil.
func (s *CompanyService) GetCompanyByCode(code string) (*models.CompanyModel, error) {
	var company models.CompanyModel
	err := s.ctx.DB.Where("code = ?", code).First(&company).Error
	return &company, err
}

// GetCompanies retrieves a paginated list of companies from the database.
//
// It takes an HTTP request and a search query string as input. The method uses
// GORM to query the database for companies, applying the search query to the
// company name and address fields. If the request contains a company ID header,
// the method also filters the result by the company ID. The function utilizes
// pagination to manage the result set and applies any necessary request
// modifications using the utils.FixRequest utility.
//
// The function returns a paginated page of CompanyModel and an error if the
// operation fails.
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

// GetCategories retrieves a list of company categories from the database.
//
// If the sector ID is not nil, the method filters the result by the sector ID.
// The function returns a slice of CompanyCategory and an error if the operation fails.
func (s *CompanyService) GetCategories(sectorID *string) ([]models.CompanyCategory, error) {
	var categories []models.CompanyCategory
	db := s.ctx.DB
	if sectorID != nil {
		db = db.Where("sector_id = ?", sectorID)
	}
	err := db.Find(&categories).Error
	return categories, err
}

// GetSectors retrieves a list of company sectors from the database.
//
// The method preloads the associated categories for each sector.
// It returns a slice of CompanySector and an error if the operation fails.

func (s *CompanyService) GetSectors() ([]models.CompanySector, error) {
	var sectors []models.CompanySector
	err := s.ctx.DB.Preload("Categories").Find(&sectors).Error
	return sectors, err
}

// GetCompaniesByUserID retrieves a list of companies associated with a specific user ID.
//
// It returns a slice of CompanyModel and an error if the operation fails.
func (s *CompanyService) GetCompaniesByUserID(userID string) ([]models.CompanyModel, error) {
	var companies []models.CompanyModel
	err := s.ctx.DB.Joins("JOIN user_companies AS uc ON companies.id = uc.company_model_id").Where("uc.user_model_id = ?", userID).Find(&companies).Error
	return companies, err
}
