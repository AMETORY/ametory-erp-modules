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

func NewDistributorService(db *gorm.DB, ctx *context.ERPContext) *DistributorService {
	return &DistributorService{db: db, ctx: ctx}
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&models.DistributorModel{})
}

type DistributorService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func (s *DistributorService) CreateDistributor(data *models.DistributorModel) error {
	return s.db.Create(data).Error
}

func (s *DistributorService) UpdateDistributor(id string, data *models.DistributorModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

func (s *DistributorService) DeleteDistributor(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.DistributorModel{}).Error
}

func (s *DistributorService) GetDistributorByID(id string) (*models.DistributorModel, error) {
	var distributor models.DistributorModel
	err := s.db.Where("id = ?", id).First(&distributor).Error

	return &distributor, err
}

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
func (s *DistributorService) GetAllProduct(request http.Request, search string, distibutorID string, status *string) (paginate.Page, error) {
	inventorySrv := s.ctx.InventoryService.(*inventory.InventoryService)
	request.Header.Set("ID-Distributor", distibutorID)
	return inventorySrv.ProductService.GetProducts(request, search, status)
}
