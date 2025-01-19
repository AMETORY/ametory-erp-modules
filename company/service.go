package company

import (
	"fmt"
	"log"
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type CompanyService struct {
	ctx *context.ERPContext
}

func NewCompanyService(ctx *context.ERPContext) *CompanyService {
	fmt.Println("INIT COMPANY SERVICE")
	var service = CompanyService{ctx: ctx}
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
	s.ctx.DB.Migrator().AlterColumn(&models.CompanyModel{}, "status")
	return s.ctx.DB.AutoMigrate(&models.CompanyModel{})
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
