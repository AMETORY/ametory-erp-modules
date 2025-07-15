package saving

import (
	"errors"
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/cooperative/cooperative_setting"
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type SavingService struct {
	db                        *gorm.DB
	ctx                       *context.ERPContext
	cooperativeSettingService *cooperative_setting.CooperativeSettingService
}

// NewSavingService creates a new instance of SavingService.
//
// The SavingService is responsible for managing saving accounts within
// a cooperative, providing functionalities such as generating account numbers
// and handling saving-related operations.
//
// It requires:
//   - db: a pointer to a GORM database instance for database operations
//   - ctx: a pointer to an ERPContext that includes user and request context
//   - cooperativeSettingService: a service to access cooperative settings

func NewSavingService(db *gorm.DB, ctx *context.ERPContext, cooperativeSettingService *cooperative_setting.CooperativeSettingService) *SavingService {
	return &SavingService{
		db:                        db,
		ctx:                       ctx,
		cooperativeSettingService: cooperativeSettingService,
	}
}

// SetDB sets the database instance for this service.
//
// This can be used to set a specific database instance for testing or
// in scenarios where the service is used with a different database than the one
// provided in the constructor.
func (s *SavingService) SetDB(db *gorm.DB) {
	s.db = db
}

// CreateTransaction creates a new transaction for the given saving.
//
// The transaction is created according to the type of saving and the
// cooperative setting. The saving amount must match the cooperative
// setting for principal and mandatory savings. For voluntary savings,
// the amount must be higher than the cooperative setting unless the
// forceVoluntry parameter is set to true.
//
// The transaction is created with a secondary transaction reference to
// the saving.
//
// The function returns an error if the saving amount does not match the
// cooperative setting, if the saving type is invalid, or if there is an
// error creating the transaction.
func (s *SavingService) CreateTransaction(saving models.SavingModel, forceVoluntry bool) error {

	if saving.Company == nil {
		return errors.New("company id is required")
	}
	var member models.CooperativeMemberModel
	if err := s.db.Model(&member).Where("id = ?", *saving.CooperativeMemberID).First(&member).Error; err != nil {
		return err
	}

	setting, err := s.cooperativeSettingService.GetSetting(saving.CompanyID)
	if err != nil {
		return err
	}

	transID := utils.Uuid()
	secTransID := utils.Uuid()

	var savingTrans = models.TransactionModel{
		BaseModel:                   shared.BaseModel{ID: transID},
		Code:                        utils.RandString(10, false),
		Date:                        *saving.Date,
		UserID:                      saving.UserID,
		CompanyID:                   saving.CompanyID,
		IsSaving:                    true,
		Credit:                      saving.Amount,
		Amount:                      saving.Amount,
		Notes:                       saving.Notes,
		SavingID:                    &saving.ID,
		CooperativeMemberID:         saving.CooperativeMemberID,
		TransactionRefID:            &secTransID,
		TransactionRefType:          "transaction",
		TransactionSecondaryRefID:   &saving.ID,
		TransactionSecondaryRefType: "cooperative-saving",
	}
	switch saving.SavingType {
	case "PRINCIPAL":
		if saving.Amount != setting.PrincipalSavingsAmount {
			return errors.New("principal savings amount must match the cooperative setting")
		}
		savingTrans.AccountID = setting.PrincipalSavingsAccountID
		savingTrans.Description = "Simpanan Pokok " + member.Name
	case "MANDATORY":
		if saving.Amount != setting.MandatorySavingsAmount {
			return errors.New("mandatory savings amount must match the cooperative setting")
		}
		savingTrans.AccountID = setting.MandatorySavingsAccountID
		savingTrans.Description = "Simpanan Wajib " + member.Name
	case "VOLUNTARY":
		if saving.Amount < setting.VoluntarySavingsAmount && !forceVoluntry {
			return errors.New("voluntary savings amount must be higher than the cooperative setting")
		}
		savingTrans.AccountID = setting.VoluntarySavingsAccountID
		savingTrans.Description = "Simpanan Sukarela " + member.Name
	default:
		return errors.New("invalid saving type")
	}

	err = s.db.Create(&savingTrans).Error
	if err != nil {
		return err
	}

	return s.db.Create(&models.TransactionModel{
		BaseModel:                   shared.BaseModel{ID: secTransID},
		Code:                        utils.RandString(10, false),
		UserID:                      saving.UserID,
		Date:                        *saving.Date,
		CompanyID:                   saving.CompanyID,
		IsSaving:                    true,
		Debit:                       saving.Amount,
		Amount:                      saving.Amount,
		AccountID:                   saving.AccountDestinationID,
		Description:                 savingTrans.Description,
		Notes:                       saving.Notes,
		SavingID:                    &saving.ID,
		CooperativeMemberID:         saving.CooperativeMemberID,
		TransactionRefID:            &transID,
		TransactionRefType:          "transaction",
		TransactionSecondaryRefID:   &saving.ID,
		TransactionSecondaryRefType: "cooperative-saving",
	}).Error
}

// GetSavings retrieves a paginated list of savings from the database.
//
// It takes an HTTP request and a search query string as parameters. The search query
// is applied to the savings description field. If a company ID is present in the
// request header, the result is filtered by the company ID. The function uses
// pagination to manage the result set and includes any necessary request
// modifications using the utils.FixRequest utility.
//
// The function returns a paginated page of SavingModel and an error if the
// operation fails.
func (s *SavingService) GetSavings(request http.Request, search string, memberID *string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.
		Preload("Company").
		Preload("User").
		Preload("Transactions", func(tx *gorm.DB) *gorm.DB {
			return tx.Preload("Account", func(tx *gorm.DB) *gorm.DB {
				return tx.Select("id", "name")
			})
		}).
		Preload("CooperativeMember").
		Preload("AccountDestination")
	if search != "" {
		stmt = stmt.Where("description ILIKE ? ",
			"%"+search+"%",
		)
	}
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ? or company_id is null", request.Header.Get("ID-Company"))
	}
	if memberID != nil {
		stmt = stmt.Where("member_id = ?", memberID)
	}
	request.URL.Query().Get("page")
	stmt = stmt.Model(&models.SavingModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.SavingModel{})
	page.Page = page.Page + 1
	return page, nil
}

// GetSavingByID retrieves a saving by its ID.
//
// It takes a string id and an optional memberID as parameters. If memberID
// is provided, the query is filtered to only include savings associated with
// the given member. The function returns a pointer to a SavingModel and an
// error if the operation fails.

func (s *SavingService) GetSavingByID(id string, memberID *string) (*models.SavingModel, error) {
	var loan models.SavingModel
	db := s.db
	if memberID != nil {
		db = db.Where("member_id = ?", memberID)
	}
	if err := db.Where("id = ?", id).First(&loan).Error; err != nil {
		return nil, err
	}
	return &loan, nil
}

// CreateSaving creates a new saving in the database.
//
// It takes a pointer to a SavingModel as a parameter and returns an error
// if the operation fails. It generates a number for the saving and
// saves it to the database.
func (c *SavingService) CreateSaving(saving *models.SavingModel) error {
	c.GenNumber(saving, saving.CompanyID)
	return c.ctx.DB.Create(saving).Error
}

// UpdateSaving updates an existing saving in the database.
//
// It takes the ID of the saving and a pointer to a SavingModel as parameters.
// It returns an error if the operation fails.
func (c *SavingService) UpdateSaving(id string, loan *models.SavingModel) error {
	return c.ctx.DB.Where("id = ?", id).Save(loan).Error
}

// DeleteSaving deletes a saving and its related transactions from the database.
//
// It takes a string id as a parameter and returns an error if the operation fails.
// The function first deletes all transactions related to the saving,
// then deletes the saving itself.
func (c *SavingService) DeleteSaving(id string) error {
	err := c.ctx.DB.Where("saving_id = ?", id).Delete(&models.TransactionModel{}).Error
	if err != nil {
		return err
	}
	return c.ctx.DB.Delete(&models.SavingModel{}, "id = ?", id).Error
}

// GenNumber generates the next number for a new saving. It queries the database to get the latest
// saving number for the given company, and then uses the invoice bill setting to generate the next
// number. If the query fails, it falls back to generating the number from the invoice bill setting
// with a prefix of "00".
func (c *SavingService) GenNumber(saving *models.SavingModel, companyID *string) error {
	setting, err := c.cooperativeSettingService.GetSetting(companyID)
	if err != nil {
		return err
	}
	lastLoan := models.SavingModel{}
	nextNumber := ""
	data := shared.InvoiceBillSettingModel{
		StaticCharacter:       setting.SavingStaticCharacter,
		NumberFormat:          setting.NumberFormat,
		AutoNumericLength:     setting.AutoNumericLength,
		RandomNumericLength:   setting.RandomNumericLength,
		RandomCharacterLength: setting.RandomCharacterLength,
	}
	if err := c.db.Where("company_id = ?", companyID).Limit(1).Order("created_at desc").Find(&lastLoan).Error; err != nil {
		nextNumber = shared.GenerateInvoiceBillNumber(data, "00")
	} else {
		nextNumber = shared.ExtractNumber(data, lastLoan.SavingNumber)
	}

	saving.SavingNumber = nextNumber
	return nil
}
