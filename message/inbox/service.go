package inbox

import (
	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"gorm.io/gorm"
)

type InboxService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewInboxService(db *gorm.DB, ctx *context.ERPContext) *InboxService {
	return &InboxService{db: db, ctx: ctx}
}

func (s *InboxService) SendMessage(data *models.InboxMessageModel) error {
	var inbox models.InboxModel
	if data.RecipientMemberID != nil {
		if err := s.db.Where("member_id = ?", data.RecipientMemberID).First(&inbox).Error; err == nil {
			data.InboxID = &inbox.ID
		} else {
			inbox.MemberID = data.RecipientMemberID
			inbox.IsDefault = true
			err := s.db.Create(&inbox).Error
			if err != nil {
				return err
			}
			data.InboxID = &inbox.ID
		}
	}
	if data.RecipientUserID != nil {
		if err := s.db.Where("user_id = ?", data.RecipientUserID).First(&inbox).Error; err == nil {
			data.InboxID = &inbox.ID
		} else {
			inbox.UserID = data.RecipientUserID
			inbox.IsDefault = true
			err := s.db.Create(&inbox).Error
			if err != nil {
				return err
			}
			data.InboxID = &inbox.ID
		}
	}

	if err := s.db.Create(&data).Error; err != nil {
		return err
	}
	return nil
}
