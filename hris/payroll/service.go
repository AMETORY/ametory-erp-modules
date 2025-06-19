package payroll

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type PayrollService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewPayrollService(ctx *context.ERPContext) *PayrollService {
	return &PayrollService{db: ctx.DB, ctx: ctx}
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.PayRollModel{},
		&models.PayrollItemModel{},
		&models.PayRollCostModel{},
		&models.PayRollInstallment{},
		&models.PayRollPeriodeModel{},
	)
}

func (s *PayrollService) CreatePayRoll(payRoll *models.PayRollModel) error {
	return s.db.Create(payRoll).Error
}

func (s *PayrollService) GetPayRollByID(id string) (*models.PayRollModel, error) {
	var payRoll models.PayRollModel
	err := s.db.First(&payRoll, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &payRoll, nil
}

func (s *PayrollService) UpdatePayRoll(payRoll *models.PayRollModel) error {
	return s.db.Save(payRoll).Error
}

func (s *PayrollService) DeletePayRoll(id string) error {
	return s.db.Delete(&models.PayRollModel{}, "id = ?", id).Error
}

func (s *PayrollService) AddItemByPayroll(payRollID string, item *models.PayrollItemModel) error {
	return s.db.Model(&models.PayRollModel{}).Where("id = ?", payRollID).Association("Items").Append(item)
}

func (s *PayrollService) UpdateItemByPayroll(payRollID string, item *models.PayrollItemModel) error {
	return s.db.Model(&models.PayRollModel{}).Where("id = ?", payRollID).Association("Items").Replace(item)
}

func (s *PayrollService) DeleteItemByPayroll(payRollID string, item *models.PayrollItemModel) error {
	return s.db.Model(&models.PayRollModel{}).Where("id = ?", payRollID).Association("Items").Delete(item)
}

func (s *PayrollService) FindAllPayroll(request *http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.Model(&models.PayRollModel{})
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.PayRollModel{})
	page.Page = page.Page + 1
	return page, nil
}

func (s *PayrollService) GetItemsFromPayroll(payRollID string) ([]*models.PayrollItemModel, error) {
	var items []*models.PayrollItemModel
	err := s.db.Model(&models.PayRollModel{}).Where("id = ?", payRollID).Association("Items").Find(&items)
	if err != nil {
		return nil, err
	}
	return items, nil
}
