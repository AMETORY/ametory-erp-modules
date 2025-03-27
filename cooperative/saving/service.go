package saving

import (
	"errors"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/cooperative/cooperative_setting"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
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
	if err := s.db.Model(&member).Where("uu_id = ?", *saving.MemberID).First(&member).Error; err != nil {
		return err
	}

	setting, err := s.cooperativeSettingService.GetSetting(saving.CompanyID)
	if err != nil {
		return err
	}

	var savingTrans = models.TransactionModel{
		Date:                *saving.Date,
		UserID:              saving.UserID,
		CompanyID:           saving.CompanyID,
		IsSaving:            true,
		Debit:               saving.Amount,
		Notes:               saving.Notes,
		SavingID:            &saving.ID,
		CooperativeMemberID: saving.MemberID,
	}
	if saving.SavingType == "Principal" {
		if saving.Amount != setting.PrincipalSavingsAmount {
			return errors.New("principal savings amount must match the cooperative setting")
		}
		savingTrans.AccountID = setting.PrincipalSavingsAccountID
		savingTrans.Description = "Simpanan Pokok " + member.Name
	}
	if saving.SavingType == "Mandatory" {
		if saving.Amount != setting.MandatorySavingsAmount {
			return errors.New("mandatory savings amount must match the cooperative setting")
		}
		savingTrans.AccountID = setting.MandatorySavingsAccountID
		savingTrans.Description = "Simpanan Wajib " + member.Name
	}
	if saving.SavingType == "Voluntary" {
		if saving.Amount < setting.VoluntarySavingsAmount && !forceVoluntry {
			return errors.New("voluntary savings amount must be higher than the cooperative setting")
		}
		savingTrans.AccountID = setting.VoluntarySavingsAccountID
		savingTrans.Description = "Simpanan Sukarela " + member.Name
	}

	err = s.db.Create(&savingTrans).Error
	if err != nil {
		return err
	}

	if err != nil {
		return err
	}

	return s.db.Create(&models.TransactionModel{
		UserID:              saving.UserID,
		Date:                *saving.Date,
		CompanyID:           saving.CompanyID,
		IsSaving:            true,
		Credit:              saving.Amount,
		AccountID:           saving.AccountDestinationID,
		Description:         savingTrans.Description,
		Notes:               saving.Notes,
		SavingID:            &saving.ID,
		CooperativeMemberID: saving.MemberID,
	}).Error
}
