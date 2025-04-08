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

func NewSavingService(db *gorm.DB, ctx *context.ERPContext, cooperativeSettingService *cooperative_setting.CooperativeSettingService) *SavingService {
	return &SavingService{
		db:                        db,
		ctx:                       ctx,
		cooperativeSettingService: cooperativeSettingService,
	}
}

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
	if saving.SavingType == "PRINCIPAL" {
		if saving.Amount != setting.PrincipalSavingsAmount {
			return errors.New("principal savings amount must match the cooperative setting")
		}
		savingTrans.AccountID = setting.PrincipalSavingsAccountID
		savingTrans.Description = "Simpanan Pokok " + member.Name
	} else if saving.SavingType == "MANDATORY" {
		if saving.Amount != setting.MandatorySavingsAmount {
			return errors.New("mandatory savings amount must match the cooperative setting")
		}
		savingTrans.AccountID = setting.MandatorySavingsAccountID
		savingTrans.Description = "Simpanan Wajib " + member.Name
	} else if saving.SavingType == "VOLUNTARY" {
		if saving.Amount < setting.VoluntarySavingsAmount && !forceVoluntry {
			return errors.New("voluntary savings amount must be higher than the cooperative setting")
		}
		savingTrans.AccountID = setting.VoluntarySavingsAccountID
		savingTrans.Description = "Simpanan Sukarela " + member.Name
	} else {
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
func (c *SavingService) CreateSaving(loan *models.SavingModel) error {
	return c.ctx.DB.Create(loan).Error
}

func (c *SavingService) UpdateSaving(id string, loan *models.SavingModel) error {
	return c.ctx.DB.Where("id = ?", id).Save(loan).Error
}

func (c *SavingService) DeleteSaving(id string) error {
	err := c.ctx.DB.Where("saving_id = ?", id).Delete(&models.TransactionModel{}).Error
	if err != nil {
		return err
	}
	return c.ctx.DB.Delete(&models.SavingModel{}, "id = ?", id).Error
}
