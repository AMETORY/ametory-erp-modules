package distributor

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/inventory"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

// NewDistributorService creates a new DistributorService, which is used to manage distributors.
// This includes creating, reading, updating, and deleting distributors.
func NewDistributorService(db *gorm.DB, ctx *context.ERPContext) *DistributorService {
	return &DistributorService{db: db, ctx: ctx}
}

// Migrate migrates the database for DistributorService.
//
// It migrates the following database tables:
//
// - distributors
//
// This function does not return an error if the database migration is successful.
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&models.DistributorModel{})
}

type DistributorService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

// CreateDistributor creates a new distributor based on the given DistributorModel.
//
// This function returns an error if the creation of the distributor fails.
func (s *DistributorService) CreateDistributor(data *models.DistributorModel) error {
	return s.db.Create(data).Error
}

// UpdateDistributor updates the distributor with the given ID using the given DistributorModel.
//
// This function returns an error if the update of the distributor fails.
func (s *DistributorService) UpdateDistributor(id string, data *models.DistributorModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

// DeleteDistributor deletes the distributor with the given ID.
//
// This function returns an error if the deletion of the distributor fails.
func (s *DistributorService) DeleteDistributor(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.DistributorModel{}).Error
}

// GetDistributorByID retrieves a distributor by its ID.
//
// It takes a string id as parameter and returns a pointer to a DistributorModel
// and an error. The function uses GORM to query the distributors table for a
// record matching the given ID. If the distributor is found, it is returned
// along with a nil error. If not found, or in case of a query error, the
// function returns a non-nil error.

func (s *DistributorService) GetDistributorByID(id string) (*models.DistributorModel, error) {
	var distributor models.DistributorModel
	err := s.db.Where("id = ?", id).First(&distributor).Error

	return &distributor, err
}

// GetDistributors retrieves a paginated list of distributors from the database.
//
// It takes an HTTP request and a search query string as input. The search query
// is applied to the distributor's name and address fields. If the request contains
// a company ID header, the result is filtered by the company ID. The function
// utilizes pagination to manage the result set and applies any necessary request
// modifications using the utils.FixRequest utility.
//
// The function returns a paginated page of DistributorModel and an error if the
// operation fails.
func (s *DistributorService) GetDistributors(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db
	if search != "" {
		stmt = stmt.Where("distributors.name ILIKE ? OR distributors.address ILIKE ? ",
			"%"+search+"%",
			"%"+search+"%",
		)
	}
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	stmt = stmt.Model(&models.DistributorModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.DistributorModel{})
	page.Page = page.Page + 1
	return page, nil
}

// GetAllProduct retrieves a paginated list of products from the inventory service.
//
// It takes an HTTP request, a search query string, a distributor ID, and a product
// status as parameters. The search query is applied to the product's name and
// description fields. If the distributor ID is provided, the result is filtered by
// the distributor ID. If the product status is provided, the result is filtered by
// the product status. The function utilizes pagination to manage the result set and
// applies any necessary request modifications using the utils.FixRequest utility.
//
// The function returns a paginated page of ProductModel and an error if the operation
// fails.
func (s *DistributorService) GetAllProduct(request http.Request, search string, distibutorID string, status *string) (paginate.Page, error) {
	inventorySrv := s.ctx.InventoryService.(*inventory.InventoryService)
	request.Header.Set("ID-Distributor", distibutorID)
	return inventorySrv.ProductService.GetProducts(request, search, status)
}
