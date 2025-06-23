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

func NewBankService(db *gorm.DB, ctx *context.ERPContext) *BankService {
	return &BankService{ctx: ctx, db: db}
}

func (s *BankService) CreateBank(bank *models.BankModel) error {
	return s.db.Create(bank).Error
}

func (s *BankService) FindBankByID(id uuid.UUID) (*models.BankModel, error) {
	var bank models.BankModel
	err := s.db.First(&bank, id).Error

	return &bank, err
}

func (s *BankService) FindAllBanks(request *http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.Model(&models.BankModel{})

	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.BankModel{})
	page.Page = page.Page + 1
	return page, nil
}

func (s *BankService) UpdateBank(bank *models.BankModel) error {
	return s.db.Save(bank).Error
}

func (s *BankService) DeleteBank(id uuid.UUID) error {
	return s.db.Delete(&models.BankModel{}, id).Error
}
