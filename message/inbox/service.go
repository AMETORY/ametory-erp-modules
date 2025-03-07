package inbox

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type InboxService struct {
	db  *gorm.DB
	ctx *context.ERPContext
}

func NewInboxService(db *gorm.DB, ctx *context.ERPContext) *InboxService {
	return &InboxService{db: db, ctx: ctx}
}

func (s *InboxService) GetInboxes(userID *string, memberID *string) ([]models.InboxModel, error) {
	var inboxes []models.InboxModel
	if userID != nil {
		if err := s.db.Where("user_id = ?", userID).Find(&inboxes).Error; err != nil {
			return nil, err
		}
	} else if memberID != nil {
		if err := s.db.Where("member_id = ?", memberID).Find(&inboxes).Error; err != nil {
			return nil, err
		}
	}
	if len(inboxes) == 0 {
		var inbox models.InboxModel
		if memberID != nil {
			inbox.MemberID = memberID
		}
		if userID != nil {
			inbox.UserID = userID
		}
		inbox.IsDefault = true
		err := s.db.Create(&inbox).Error
		if err != nil {
			return nil, err
		}

		inboxes = append(inboxes, inbox)
	}
	return inboxes, nil
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

func (s *InboxService) GetMessageByInboxID(request http.Request, search string, inboxID *string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.Where("parent_inbox_message_id IS NULL")
	if inboxID != nil {
		stmt = stmt.Where("inbox_id = ?", *inboxID)
	}
	if search != "" {
		stmt = stmt.Where("subject ILIKE ? OR message ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
		)
	}

	stmt = stmt.Model(&models.InboxMessageModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.InboxMessageModel{})
	page.Page = page.Page + 1
	return page, nil
}

func (s *InboxService) GetDefaultInbox(userID *string, memberID *string) (*models.InboxModel, error) {
	var inbox models.InboxModel
	if userID != nil {
		if err := s.db.Where("user_id = ? AND is_default = ?", userID, true).First(&inbox).Error; err != nil {
			return nil, err
		}
	} else if memberID != nil {
		if err := s.db.Where("member_id = ? AND is_default = ?", memberID, true).First(&inbox).Error; err != nil {
			return nil, err
		}
	}
	return &inbox, nil
}

func (s *InboxService) CountUnread(userID *string, memberID *string) (int64, error) {
	var count int64
	if userID != nil {
		if err := s.db.Model(&models.InboxMessageModel{}).Where("inbox_id IN (SELECT id FROM inboxes WHERE user_id = ? AND is_default = ?) AND recipient_user_id = ? AND read = ?", userID, true, userID, false).Count(&count).Error; err != nil {
			return 0, err
		}
	} else if memberID != nil {
		if err := s.db.Model(&models.InboxMessageModel{}).Where("inbox_id IN (SELECT id FROM inboxes WHERE member_id = ? AND is_default = ?) AND recipient_member_id = ? AND read = ?", memberID, true, memberID, false).Count(&count).Error; err != nil {
			return 0, err
		}
	}
	return count, nil
}

func (s *InboxService) GetInboxMessageDetail(inboxMessageID string) (*models.InboxMessageModel, error) {
	var inboxMessage models.InboxMessageModel
	if err := s.db.
		Preload("SenderUser").
		Preload("SenderMember").
		Preload("RecipientUser").
		Preload("RecipientMember").
		Where("id = ?", inboxMessageID).First(&inboxMessage).Error; err != nil {
		return nil, err
	}

	replies, err := inboxMessage.LoadRecursiveChildren(s.db)
	if err != nil {
		return nil, err
	}
	inboxMessage.Replies = replies
	return &inboxMessage, nil
}
