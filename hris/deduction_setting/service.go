package deduction_setting

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type DeductionSettingService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

// NewDeductionSettingService creates a new instance of DeductionSettingService.
//
// It initializes the service with the provided ERP context, using it to
// access the database instance for performing CRUD operations on deduction settings.
//
// Parameters:
//   ctx (*context.ERPContext): The ERP context containing the database connection.
//
// Returns:
//   *DeductionSettingService: A pointer to the newly created DeductionSettingService.

func NewDeductionSettingService(ctx *context.ERPContext) *DeductionSettingService {
	return &DeductionSettingService{db: ctx.DB, ctx: ctx}
}

// Migrate creates the necessary database schema for DeductionSettingModel.
//
// It takes a GORM database instance as parameter and returns an error if the
// migration process encounters any issues.
//
// Parameters:
//
//	db (*gorm.DB): The GORM database instance.
//
// Returns:
//
//	error: An error if the migration process fails; otherwise nil.
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.DeductionSettingModel{},
	)
}

// Create adds a new DeductionSettingModel to the database.
//
// This function takes a DeductionSettingModel pointer as a parameter and
// attempts to insert it into the database using the GORM Create method.
//
// Parameters:
//   deductionSetting (*models.DeductionSettingModel): The deduction setting
//     model to be added to the database.
//
// Returns:
//   error: An error if the create operation fails; otherwise nil.

func (s *DeductionSettingService) Create(deductionSetting *models.DeductionSettingModel) error {
	return s.db.Create(deductionSetting).Error
}

// FindAll retrieves all deduction settings from the database.
//
// The function takes an HTTP request object as parameter and uses the query
// parameters to filter the deduction settings. The records are sorted by
// ID in descending order by default, but the order can be changed by
// specifying the "order" query parameter.
//
// The function returns a Page object containing the deduction settings and
// the pagination information. The Page object contains the following fields:
//
//	Records: []models.DeductionSettingModel
//	Page: int
//	PageSize: int
//	TotalPages: int
//	TotalRecords: int
//
// If the operation is not successful, the function returns an error object.
func (a *DeductionSettingService) FindAll(request *http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := a.db.Model(&models.DeductionSettingModel{})
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.DeductionSettingModel{})
	page.Page = page.Page + 1
	return page, nil
}

// FindOne retrieves a deduction setting by its ID from the database.
//
// The function takes a string ID as a parameter and attempts to find the
// corresponding deduction setting. If the deduction setting is found,
// the function returns the deduction setting model instance; otherwise,
// it returns an error indicating what went wrong.
//
// Parameters:
//   id (string): The ID of the deduction setting to retrieve.
//
// Returns:
//   *models.DeductionSettingModel: The deduction setting model instance
//     if found, or nil if not found.
//   error: An error object if the operation fails, or nil if it is successful.

func (s *DeductionSettingService) FindOne(id string) (*models.DeductionSettingModel, error) {
	deductionSetting := &models.DeductionSettingModel{}

	db := s.db.Model(&models.DeductionSettingModel{})

	if err := db.Where("id = ?", id).First(deductionSetting).Error; err != nil {
		return nil, err
	}

	return deductionSetting, nil
}

// Update updates an existing deduction setting in the database.
//
// The function takes a pointer to a DeductionSettingModel as a parameter and
// attempts to update the corresponding deduction setting in the database.
// If the operation is successful, the function returns nil; otherwise, it
// returns an error indicating what went wrong.
//
// Parameters:
//
//	deductionSetting (*models.DeductionSettingModel): The deduction setting
//	  model instance to be updated.
//
// Returns:
//
//	error: An error object if the update operation fails, or nil if it is
//	  successful.
func (s *DeductionSettingService) Update(deductionSetting *models.DeductionSettingModel) error {
	return s.db.Save(deductionSetting).Error
}

// Delete deletes a deduction setting from the database.
//
// The function takes the ID of the deduction setting as a parameter and
// attempts to delete it from the database. If the deletion is successful,
// the function returns nil; otherwise, it returns an error object indicating
// what went wrong.
//
// Parameters:
//
//	id (string): The ID of the deduction setting to be deleted.
//
// Returns:
//
//	error: An error object if the deletion fails, or nil if successful.
func (s *DeductionSettingService) Delete(id string) error {
	return s.db.Delete(&models.DeductionSettingModel{}, "id = ?", id).Error
}
