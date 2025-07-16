package unit

import (
	"errors"
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type UnitService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

// NewUnitService creates a new instance of UnitService with the given database connection and context.
func NewUnitService(db *gorm.DB, ctx *context.ERPContext) *UnitService {
	return &UnitService{db: db, ctx: ctx}
}

// Migrate runs database migrations for the unit module.
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.UnitModel{},
		&models.ProductUnits{},
	)
}

// CreateUnit adds a new unit to the database.
//
// The function takes a pointer to a UnitModel, which contains the data
// for the new unit. It returns an error if the creation
// of the unit fails.

func (s *UnitService) CreateUnit(data *models.UnitModel) error {
	return s.db.Create(data).Error
}

// UpdateUnit updates an existing unit in the database.
//
// The function takes a unit ID and a pointer to a UnitModel, which contains the data
// for the updated unit. It returns an error if the update
// of the unit fails.
func (s *UnitService) UpdateUnit(id string, data *models.UnitModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

// DeleteUnit deletes an existing unit from the database.
//
// The function takes a unit ID and deletes the associated unit
// from the database. It returns an error if the deletion
// of the unit fails.
func (s *UnitService) DeleteUnit(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.UnitModel{}).Error
}

// GetUnitByID retrieves a unit by its ID.
//
// The function takes a unit ID as a string and returns a pointer to a UnitModel
// containing the unit details. It returns an error if the retrieval operation fails.
func (s *UnitService) GetUnitByID(id string) (*models.UnitModel, error) {
	var invoice models.UnitModel
	err := s.db.Where("id = ?", id).First(&invoice).Error
	return &invoice, err
}

// GetUnits retrieves a paginated list of units from the database.
//
// It takes an HTTP request and a search query string as input. The method uses
// GORM to query the database for units, applying the search query to the name,
// code, and description fields. If the request contains a company ID header,
// the method filters the result by the company ID or includes entries with a
// null company ID. The function utilizes pagination to manage the result set
// and applies any necessary request modifications using the utils.FixRequest
// utility.
//
// The function returns a paginated page of UnitModel and an error if the
// operation fails.
func (s *UnitService) GetUnits(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db
	if search != "" {
		stmt = stmt.Where("name ILIKE ? OR code ILIKE ? OR description ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
			"%"+search+"%",
		)
	}
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ? or company_id is null", request.Header.Get("ID-Company"))
	}
	request.URL.Query().Get("page")
	stmt = stmt.Model(&models.UnitModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.UnitModel{})
	page.Page = page.Page + 1
	return page, nil
}

// GetUnitByName retrieves a unit from the database by its name.
//
// It takes a name as input and returns a pointer to a UnitModel and an error.
// If the unit with the given name is not found, the function creates a new
// unit with that name and saves it in the database. If the operation is
// successful, the error is nil. Otherwise, the error contains information about
// what went wrong.
func (s *UnitService) GetUnitByName(name string) (*models.UnitModel, error) {
	var brand models.UnitModel
	err := s.db.Where("name = ?", name).First(&brand).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			brand = models.UnitModel{
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
