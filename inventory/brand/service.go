package brand

import (
	"errors"
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type BrandService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

// NewBrandService creates a new BrandService with the given database and ERPContext.
//
// This function will return a new BrandService object with the given database and ERPContext.
// This object can then be used to access the BrandModel database table.
func NewBrandService(db *gorm.DB, ctx *context.ERPContext) *BrandService {
	return &BrandService{db: db, ctx: ctx}
}

// Migrate migrates the BrandModel database table.
//
// This function will migrate the BrandModel table if it does not already exist.
// If the table already exists, this function will not change the table in any
// way.
//
// This function will return an error if the migration fails.
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.BrandModel{},
	)
}

// CreateBrand creates a new brand in the database.
//
// The function takes a pointer to a BrandModel and returns an error.
// The error will be nil if the brand was created successfully.
func (s *BrandService) CreateBrand(data *models.BrandModel) error {
	return s.db.Create(data).Error
}

// UpdateBrand updates an existing brand in the database.
//
// It takes an ID and a pointer to a BrandModel as inputs and returns an error.
// The function uses GORM to update the brand data in the database where the
// brand ID matches. If the update is successful, the error is nil. Otherwise,
// the error contains information about what went wrong.

func (s *BrandService) UpdateBrand(id string, data *models.BrandModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

// DeleteBrand deletes an existing brand from the database.
//
// It takes an ID as input and returns an error. The function uses GORM to
// delete the brand data from the database where the brand ID matches. If
// the deletion is successful, the error is nil. Otherwise, the error contains
// information about what went wrong.
func (s *BrandService) DeleteBrand(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.BrandModel{}).Error
}

// GetBrandByID retrieves a brand from the database by ID.
//
// It takes an ID as input and returns a pointer to a BrandModel and an error.
// The function uses GORM to retrieve the brand data from the brands table.
// If the operation fails, an error is returned. Otherwise, the error is nil.
func (s *BrandService) GetBrandByID(id string) (*models.BrandModel, error) {
	var invoice models.BrandModel
	err := s.db.Where("id = ?", id).First(&invoice).Error
	return &invoice, err
}

// GetBrandByCode retrieves a brand by its code.
//
// It takes a code as input and returns a pointer to a BrandModel and an error.
// The function uses GORM to retrieve the brand data from the brands table.
// If the operation fails, an error is returned. Otherwise, the error is nil.
func (s *BrandService) GetBrandByCode(code string) (*models.BrandModel, error) {
	var invoice models.BrandModel
	err := s.db.Where("code = ?", code).First(&invoice).Error
	return &invoice, err
}

// GetBrands retrieves a paginated list of brands from the database.
//
// It takes an http.Request and a search query string as input. The method uses
// GORM to query the database for brands, applying the search query to the
// brand name and description fields. If the request contains a company ID
// header, the method also filters the result by the company ID. The function
// utilizes pagination to manage the result set and applies any necessary request
// modifications using the utils.FixRequest utility.
//
// The function returns a paginated page of BrandModel and an error if the
// operation fails.
func (s *BrandService) GetBrands(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db
	if search != "" {
		stmt = stmt.Where("brands.description ILIKE ? OR brands.name ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
		)
	}
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	request.URL.Query().Get("page")
	stmt = stmt.Model(&models.BrandModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.BrandModel{})
	page.Page = page.Page + 1
	return page, nil
}

// GetBrandByName retrieves a brand from the database by its name.
//
// It takes a name as input and returns a pointer to a BrandModel and an error.
// If the brand with the given name is not found, the function creates a new
// brand with that name and saves it in the database. If the operation is
// successful, the error is nil. Otherwise, the error contains information about
// what went wrong.
func (s *BrandService) GetBrandByName(name string) (*models.BrandModel, error) {
	var brand models.BrandModel
	err := s.db.Where("name = ?", name).First(&brand).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			brand = models.BrandModel{
				Name: name,
			}
			err := s.db.Create(&brand).Error
			if err != nil {
				return nil, err
			}
			return &brand, nil
		}
		return nil, err
	}
	return &brand, nil
}
