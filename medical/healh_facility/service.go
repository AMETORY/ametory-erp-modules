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

// NewHeathFacilityService creates a new instance of HeathFacilityService with the given database connection and context.
func NewHeathFacilityService(db *gorm.DB, ctx *context.ERPContext) *HeathFacilityService {
	return &HeathFacilityService{db: db, ctx: ctx}
}

// CreateFacility creates a new health facility.
//
// It takes a HealthFacilityModel instance as an argument and returns an error if
// the creation fails.
func (s *HeathFacilityService) CreateFacility(input models.HealthFacilityModel) error {
	return s.db.Create(&input).Error
}

// UpdateFacility updates an existing health facility.
//
// It takes the ID of the health facility as its first argument and a HealthFacilityModel
// instance as its second argument. It returns an error if the update fails.
func (s *HeathFacilityService) UpdateFacility(id string, input models.HealthFacilityModel) error {
	return s.db.Where("id = ?", id).Updates(&input).Error
}

// DeleteFacility deletes a health facility.
//
// It takes the ID of the health facility as its argument and returns an error if
// the deletion fails.
func (s *HeathFacilityService) DeleteFacility(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.HealthFacilityModel{}).Error
}

// GetFacilityDetail retrieves a health facility by its ID.
//
// It takes the ID of the health facility as its argument and returns the
// corresponding HealthFacilityModel instance if the retrieval is successful. If
// the retrieval fails, it returns an error.
func (s *HeathFacilityService) GetFacilityDetail(id string) (*models.HealthFacilityModel, error) {
	var result models.HealthFacilityModel
	err := s.db.Where("id = ?", id).First(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetFacilities retrieves a list of health facilities.
//
// It takes an HTTP request and a search string as its arguments. The search string
// is used to filter the results by name or description. It returns a paginate.Page
// instance if the retrieval is successful. If the retrieval fails, it returns an
// error.
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

// CreateSubFacility creates a new sub-facility.
//
// It takes a SubFacilityModel instance as its argument and returns an error if the
// creation fails.
func (s *HeathFacilityService) CreateSubFacility(input models.SubFacilityModel) error {
	return s.db.Create(&input).Error
}

// UpdateSubFacility updates an existing sub-facility.
//
// It takes the ID of the sub-facility as its first argument and a SubFacilityModel
// instance as its second argument. It returns an error if the update fails.
func (s *HeathFacilityService) UpdateSubFacility(id string, input models.SubFacilityModel) error {
	return s.db.Where("id = ?", id).Updates(&input).Error
}

// DeleteSubFacility deletes a sub-facility.
//
// It takes the ID of the sub-facility as its argument and returns an error if the
// deletion fails.
func (s *HeathFacilityService) DeleteSubFacility(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.SubFacilityModel{}).Error
}

// GetSubFacilityDetail retrieves a sub-facility by its ID.
//
// It takes the ID of the sub-facility as its argument and returns the corresponding
// SubFacilityModel instance if the retrieval is successful. If the retrieval fails,
// it returns an error.
func (s *HeathFacilityService) GetSubFacilityDetail(id string) (*models.SubFacilityModel, error) {
	var result models.SubFacilityModel
	err := s.db.Where("id = ?", id).First(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetSubFacilities retrieves a list of sub-facilities.
//
// It takes an HTTP request and a search string as its arguments. The search string
// is used to filter the results by name or phone number. It returns a paginate.Page
// instance if the retrieval is successful. If the retrieval fails, it returns an
// error.
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
