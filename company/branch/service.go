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

func NewBranchService(db *gorm.DB, ctx *context.ERPContext) *BranchService {
	return &BranchService{db: db, ctx: ctx}
}

func (s *BranchService) CreateBranch(data *models.BranchModel) error {
	return s.db.Create(data).Error
}

func (s *BranchService) UpdateBranch(id string, data *models.BranchModel) error {
	return s.db.Where("id = ?", id).Updates(data).Error
}

func (s *BranchService) DeleteBranch(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.BranchModel{}).Error
}

func (s *BranchService) GetBranchByID(id string) (*models.BranchModel, error) {
	var branch models.BranchModel
	if err := s.db.Where("id = ?", id).First(&branch).Error; err != nil {
		return nil, err
	}
	return &branch, nil
}

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
