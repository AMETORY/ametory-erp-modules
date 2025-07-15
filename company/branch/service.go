package branch

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type BranchService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

// NewBranchService creates a new instance of BranchService.
//
// The BranchService is responsible for interacting with the BranchModel on the database.
func NewBranchService(db *gorm.DB, ctx *context.ERPContext) *BranchService {
	return &BranchService{db: db, ctx: ctx}
}

// CreateBranch creates a new branch in the database.
//
// The function takes a pointer to a BranchModel as input and returns an error.
// If the operation is successful, the error is nil. Otherwise, the error
// contains information about what went wrong.
func (s *BranchService) CreateBranch(data *models.BranchModel) error {
	return s.db.Create(data).Error
}

// UpdateBranch updates an existing branch in the database.
//
// It takes an ID and a pointer to a BranchModel as inputs and returns an error.
// The function uses GORM to update the branch data in the database where the branch ID matches.
// If the update is successful, the error is nil. Otherwise, the error contains information about what went wrong.

func (s *BranchService) UpdateBranch(id string, data *models.BranchModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

// DeleteBranch deletes an existing branch from the database.
//
// It takes an ID as input and returns an error. The function uses GORM to
// delete the branch data from the database where the branch ID matches. If
// the deletion is successful, the error is nil. Otherwise, the error contains
// information about what went wrong.
func (s *BranchService) DeleteBranch(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.BranchModel{}).Error
}

// GetBranchByID retrieves a branch from the database by ID.
//
// It takes an ID as input and returns a pointer to a BranchModel and an error.
// The function uses GORM to retrieve the branch data from the branches table.
// If the operation fails, an error is returned.
func (s *BranchService) GetBranchByID(id string) (*models.BranchModel, error) {
	var branch models.BranchModel
	if err := s.db.Where("id = ?", id).First(&branch).Error; err != nil {
		return nil, err
	}
	return &branch, nil
}

// FindAllBranches retrieves a paginated list of branches associated with a specific company.
//
// It takes an HTTP request as input and returns a paginated Page of
// BranchModel and an error if the operation fails. The function applies a
// filter based on the company ID provided in the request header to further
// filter the branches.

func (s *BranchService) FindAllBranches(request *http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.Model(&models.BranchModel{})
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.BranchModel{})
	page.Page = page.Page + 1
	return page, nil
}
