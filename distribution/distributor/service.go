package distributor

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

func NewDistributorService(db *gorm.DB, ctx *context.ERPContext) *DistributorService {
	return &DistributorService{db: db, ctx: ctx}
}

type DistributorService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func (s *DistributorService) CreateDistributor(data *DistributorModel) error {
	return s.db.Create(data).Error
}

func (s *DistributorService) UpdateDistributor(id string, data *DistributorModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

func (s *DistributorService) DeleteDistributor(id string) error {
	return s.db.Where("id = ?", id).Delete(&DistributorModel{}).Error
}

func (s *DistributorService) GetDistributorByID(id string) (*DistributorModel, error) {
	var distributor DistributorModel
	err := s.db.Where("id = ?", id).First(&distributor).Error

	return &distributor, err
}

func (s *DistributorService) GetDistributors(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db
	if search != "" {
		stmt = stmt.Where("distributors.name LIKE ? OR distributors.address LIKE ? ",
			"%"+search+"%",
			"%"+search+"%",
		)
	}
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	stmt = stmt.Model(&DistributorModel{})
	page := pg.With(stmt).Request(request).Response(&[]DistributorModel{})
	return page, nil
}
