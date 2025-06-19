package payroll

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
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
