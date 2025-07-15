package bank

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/google/uuid"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type BankService struct {
	ctx *context.ERPContext
	db  *gorm.DB
}

// NewBankService returns a new instance of BankService.
//
// The service is created by providing a GORM database instance and an ERP context.
// The ERP context is used for authentication and authorization purposes, while the
// database instance is used for CRUD (Create, Read, Update, Delete) operations.

func NewBankService(db *gorm.DB, ctx *context.ERPContext) *BankService {
	return &BankService{ctx: ctx, db: db}
}

// CreateBank creates a new bank in the database.
//
// The method takes a pointer to a BankModel and returns an error.
func (s *BankService) CreateBank(bank *models.BankModel) error {
	return s.db.Create(bank).Error
}

// FindBankByID returns a bank by its ID.
//
// The method takes a UUID bank ID, queries the database and returns a pointer to
// a BankModel and an error. If the bank is found, the error is nil. Otherwise, the
// error is not nil and the BankModel pointer is nil.
func (s *BankService) FindBankByID(id uuid.UUID) (*models.BankModel, error) {
	var bank models.BankModel
	err := s.db.First(&bank, id).Error

	return &bank, err
}

// FindAllBanks returns a paginated list of all banks in the database.
//
// The method takes an HTTP request and returns a paginated page of BankModel
// objects and an error. The request is used for pagination purposes, such as
// specifying the page number and the number of items per page.
//
// If the method successfully retrieves the list of banks from the database,
// the error is nil. Otherwise, the error is not nil and the paginated page is
// empty.
func (s *BankService) FindAllBanks(request *http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.Model(&models.BankModel{})

	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.BankModel{})
	page.Page = page.Page + 1
	return page, nil
}

// UpdateBank updates an existing bank in the database.
//
// The function takes a pointer to a BankModel as its argument. The BankModel
// instance contains the new values to be updated in the database.
//
// The function returns an error if the update fails.
func (s *BankService) UpdateBank(bank *models.BankModel) error {
	return s.db.Save(bank).Error
}

// DeleteBank deletes a bank from the database by its ID.
//
// The function takes a UUID bank ID as its argument and returns an error.
// It uses GORM to perform the deletion operation on the BankModel table.
// If the deletion is successful, the error is nil. Otherwise, the error contains
// information about what went wrong.

func (s *BankService) DeleteBank(id uuid.UUID) error {
	return s.db.Delete(&models.BankModel{}, id).Error
}
