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

func NewTaxService(db *gorm.DB, ctx *context.ERPContext, accountService *account.AccountService) *TaxService {
	return &TaxService{
		db:             db,
		ctx:            ctx,
		accountService: accountService,
	}
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&models.TaxModel{})
}

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

func (ts *TaxService) GetTaxByID(id string) (*models.TaxModel, error) {
	var taxModel models.TaxModel

	if err := ts.db.Where("id = ?", id).First(&taxModel).Error; err != nil {
		return nil, err
	}

	return &taxModel, nil
}

func (ts *TaxService) CreateTax(taxModel *models.TaxModel) error {

	return ts.db.Create(taxModel).Error
}

func (ts *TaxService) UpdateTax(id string, taxModel *models.TaxModel) error {

	return ts.db.Where("id = ?", id).Updates(taxModel).Error
}

func (ts *TaxService) DeleteTax(id string) error {
	if err := ts.db.Where("id = ?", id).Delete(&models.TaxModel{}).Error; err != nil {
		return err
	}

	return nil
}
