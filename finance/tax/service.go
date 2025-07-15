package tax

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/finance/account"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type TaxService struct {
	db             *gorm.DB
	ctx            *context.ERPContext
	accountService *account.AccountService
}

// NewTaxService returns a new instance of TaxService.
//
// The service is created by providing a GORM database instance, an ERP context, and an
// AccountService instance. The ERP context is used for authentication and authorization
// purposes, while the database instance is used for CRUD (Create, Read, Update, Delete)
// operations. The AccountService is used for retrieving the account information for taxes.
func NewTaxService(db *gorm.DB, ctx *context.ERPContext, accountService *account.AccountService) *TaxService {
	return &TaxService{
		db:             db,
		ctx:            ctx,
		accountService: accountService,
	}
}

// Migrate runs the database migration for the TaxModel.
//
// The method takes a GORM database instance and returns an error if the migration
// fails. The migration is required to create the tax table in the database.
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&models.TaxModel{})
}

// GetTaxes returns a paginated page of TaxModel, with optional search query.
//
// The method takes an http.Request and a search string as parameters. The search
// string is used to query the tax table in the database. The query is done by
// searching for the search string in the name and code of the tax model. The
// search is case-insensitive and uses the ILIKE operator.
//
// The method returns a paginate.Page object, which contains the result of the
// query and the pagination information. The result is stored in the Response
// field of the paginate.Page object.
//
// The method returns an error if there is an error executing the query or
// validating the request.
func (ts *TaxService) GetTaxes(request *http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := ts.db.Preload("AccountReceivable").Preload("AccountPayable")
	if search != "" {
		stmt = stmt.Where("name ILIKE ? OR code ILIKE ?", "%"+search+"%", "%"+search+"%")
	}
	stmt = stmt.Model(&models.TaxModel{})
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.TaxModel{})
	page.Page = page.Page + 1
	return page, nil
}

// GetTaxByID returns a TaxModel object by ID.
//
// The method takes a string id as the parameter and returns a TaxModel
// object and an error. The error is returned if the query fails.
//
// The method uses the GORM First method to query the tax table in the
// database. The query is done by searching for the tax id in the id field.
//
// The method returns the TaxModel object and a nil error if the query is
// successful. Otherwise, it returns a nil TaxModel and an error.
func (ts *TaxService) GetTaxByID(id string) (*models.TaxModel, error) {
	var taxModel models.TaxModel

	if err := ts.db.Where("id = ?", id).First(&taxModel).Error; err != nil {
		return nil, err
	}

	return &taxModel, nil
}

// CreateTax creates a new tax record in the database.
//
// The method takes a pointer to a TaxModel object as a parameter and
// attempts to insert it into the tax table. It returns an error if the
// creation operation fails.

func (ts *TaxService) CreateTax(taxModel *models.TaxModel) error {

	return ts.db.Create(taxModel).Error
}

// UpdateTax updates an existing tax record in the database.
//
// The method takes a string id and a pointer to a TaxModel object as parameters.
// It uses the GORM Updates method to modify the tax record in the database where
// the id matches the provided id. The TaxModel object contains the new values
// for the update.
//
// The method returns an error if the update operation fails. If the update is
// successful, the error is nil.

func (ts *TaxService) UpdateTax(id string, taxModel *models.TaxModel) error {

	return ts.db.Where("id = ?", id).Updates(taxModel).Error
}

// DeleteTax deletes a tax record from the database.
//
// The method takes a string id as a parameter and uses the GORM Delete method to
// delete the tax record from the database where the id matches the provided id.
//
// The method returns an error if the deletion operation fails. If the deletion is
// successful, the error is nil.
func (ts *TaxService) DeleteTax(id string) error {
	if err := ts.db.Where("id = ?", id).Delete(&models.TaxModel{}).Error; err != nil {
		return err
	}

	return nil
}
