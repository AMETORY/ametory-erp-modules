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
	if userID != nil && memberID != nil {
		if err := s.db.Where("user_id = ? AND member_id = ?", userID, memberID).Find(&inboxes).Error; err != nil {
			return nil, err
		}
	} else if userID != nil {
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
		var trash models.InboxModel
		if memberID != nil {
			trash.MemberID = memberID
		}
		if userID != nil {
			trash.UserID = userID
		}
		trash.Name = "TRASH"
		trash.IsTrash = true
		err = s.db.Create(&trash).Error
		if err != nil {
			return nil, err
		}

		inboxes = append(inboxes, trash)
	}
	var isTrashExist = false
	for _, v := range inboxes {
		if v.IsTrash {
			isTrashExist = true
		}
	}

	if !isTrashExist {
		var trash models.InboxModel
		if memberID != nil {
			trash.MemberID = memberID
		}
		if userID != nil {
			trash.UserID = userID
		}
		trash.Name = "TRASH"
		trash.IsTrash = true
		err := s.db.Create(&trash).Error
		if err != nil {
			return nil, err
		}
		inboxes = append(inboxes, trash)
	}
	return inboxes, nil
}

func (s *InboxService) SendMessage(data *models.InboxMessageModel) error {

	var inbox models.InboxModel
	if data.RecipientMemberID != nil {
		if err := s.db.Where("member_id = ? AND is_default = ?", data.RecipientMemberID, true).First(&inbox).Error; err == nil {
			data.InboxID = &inbox.ID
		} else {
			var member models.MemberModel
			if err := s.db.Where("id = ?", data.RecipientMemberID).First(&member).Error; err != nil {
				return err
			}
			inbox.MemberID = data.RecipientMemberID
			inbox.UserID = &member.UserID
			inbox.IsDefault = true
			err := s.db.Create(&inbox).Error
			if err != nil {
				return err
			}
			data.InboxID = &inbox.ID
		}
	}
	if data.RecipientUserID != nil {
		if err := s.db.Where("user_id = ? AND is_default = ?", data.RecipientUserID, true).First(&inbox).Error; err == nil {
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
	stmt := s.db.
		Preload("SenderUser").
		Preload("SenderMember.User").
		Preload("RecipientUser").
		Preload("RecipientMember.User").
		Where("parent_inbox_message_id IS NULL")
	if inboxID != nil {
		stmt = stmt.Where("inbox_id = ?", *inboxID)
	}
	if search != "" {
		stmt = stmt.Where("subject ILIKE ? OR message ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
		)
	}

	stmt = stmt.Order("created_at desc")

	stmt = stmt.Model(&models.InboxMessageModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.InboxMessageModel{})
	page.Page = page.Page + 1
	return page, nil
}
func (s *InboxService) GetSentMessages(request http.Request, search string, userID *string, memberID *string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.Where("parent_inbox_message_id IS NULL")
	if userID != nil && memberID != nil {
		stmt = stmt.Where("sender_user_id = ? AND sender_member_id = ?", userID, memberID)
	} else if userID != nil {
		stmt = stmt.Where("sender_user_id = ?", userID)
	} else if memberID != nil {
		stmt = stmt.Where(" sender_member_id = ?", memberID)
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
		if err := s.db.Model(&models.InboxMessageModel{}).Where("inbox_id IN (SELECT id FROM inbox WHERE user_id = ? AND is_default = ?) AND recipient_user_id = ? AND read = ?", userID, true, userID, false).Count(&count).Error; err != nil {
			return 0, err
		}
	} else if memberID != nil {
		if err := s.db.Model(&models.InboxMessageModel{}).Where("inbox_id IN (SELECT id FROM inbox WHERE member_id = ? AND is_default = ?) AND recipient_member_id = ? AND read = ?", memberID, true, memberID, false).Count(&count).Error; err != nil {
			return 0, err
		}
	}
	return count, nil
}

func (s *InboxService) DeleteMessage(inboxMessageID string, userID *string, memberID *string) error {
	var inbox *models.InboxModel
	if userID != nil && memberID != nil {
		if err := s.db.Where("user_id = ? AND member_id = ? AND is_trash = ?", userID, memberID, true).First(&inbox).Error; err != nil {
			return err
		}
	} else if userID != nil {
		if err := s.db.Where("user_id = ? AND is_trash = ?", userID, true).First(&inbox).Error; err != nil {
			return err
		}
	} else if memberID != nil {
		if err := s.db.Where("member_id = ? AND is_trash = ?", memberID, true).First(&inbox).Error; err != nil {
			return err
		}
	}

	var inboxMessage models.InboxMessageModel
	if err := s.db.Where("id = ?", inboxMessageID).First(&inboxMessage).Error; err != nil {
		return err
	}

	inboxMessage.InboxID = &inbox.ID

	return s.db.Save(&inboxMessage).Error
}

func (s *InboxService) GetInboxMessageDetail(inboxMessageID string) (*models.InboxMessageModel, error) {
	var inboxMessage models.InboxMessageModel
	if err := s.db.
		Preload("SenderUser").
		Preload("SenderMember.User").
		Preload("RecipientUser").
		Preload("RecipientMember.User").
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
