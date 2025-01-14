package company

import (
	"fmt"
	"log"
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
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

func (s *CompanyService) CreateCompany(data *CompanyModel) error {
	return s.ctx.DB.Create(data).Error
}

func (s *CompanyService) UpdateCompany(id string, data *CompanyModel) error {
	return s.ctx.DB.Where("id = ?", id).Updates(data).Error
}

func (s *CompanyService) DeleteCompany(id string) error {
	return s.ctx.DB.Where("id = ?", id).Delete(&CompanyModel{}).Error
}

func (s *CompanyService) GetCompanyByID(id string) (*CompanyModel, error) {
	var invoice CompanyModel
	err := s.ctx.DB.Where("id = ?", id).First(&invoice).Error
	return &invoice, err
}

func (s *CompanyService) GetCompanyByCode(code string) (*CompanyModel, error) {
	var invoice CompanyModel
	err := s.ctx.DB.Where("code = ?", code).First(&invoice).Error
	return &invoice, err
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
	stmt = stmt.Model(&CompanyModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]CompanyModel{})
	page.Page = page.Page + 1
	return page, nil
}
