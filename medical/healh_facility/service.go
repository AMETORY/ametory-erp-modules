package healh_facility

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type HeathFacilityService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewHeathFacilityService(db *gorm.DB, ctx *context.ERPContext) *HeathFacilityService {
	return &HeathFacilityService{db: db, ctx: ctx}
}

func (s *HeathFacilityService) CreateFacility(input models.HealthFacilityModel) error {
	return s.db.Create(&input).Error
}

func (s *HeathFacilityService) UpdateFacility(id string, input models.HealthFacilityModel) error {
	return s.db.Where("id = ?", id).Updates(&input).Error
}

func (s *HeathFacilityService) DeleteFacility(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.HealthFacilityModel{}).Error
}

func (s *HeathFacilityService) GetFacilityDetail(id string) (*models.HealthFacilityModel, error) {
	var result models.HealthFacilityModel
	err := s.db.Where("id = ?", id).First(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *HeathFacilityService) GetFacilities(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db
	if search != "" {
		stmt = stmt.Where("brands.description ILIKE ? OR brands.name ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
		)
	}

	request.URL.Query().Get("page")
	stmt = stmt.Model(&models.HealthFacilityModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.BrandModel{})
	page.Page = page.Page + 1
	return page, nil
}

func (s *HeathFacilityService) CreateSubFacility(input models.SubFacilityModel) error {
	return s.db.Create(&input).Error
}

func (s *HeathFacilityService) UpdateSubFacility(id string, input models.SubFacilityModel) error {
	return s.db.Where("id = ?", id).Updates(&input).Error
}

func (s *HeathFacilityService) DeleteSubFacility(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.SubFacilityModel{}).Error
}

func (s *HeathFacilityService) GetSubFacilityDetail(id string) (*models.SubFacilityModel, error) {
	var result models.SubFacilityModel
	err := s.db.Where("id = ?", id).First(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *HeathFacilityService) GetSubFacilities(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db
	if search != "" {
		stmt = stmt.Where("sub_facilities.name ILIKE ? OR sub_facilities.phone_number ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
		)
	}

	stmt = stmt.Model(&models.SubFacilityModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.SubFacilityModel{})
	page.Page = page.Page + 1
	return page, nil
}
