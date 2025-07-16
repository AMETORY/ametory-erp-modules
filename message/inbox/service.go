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

// NewInboxService creates a new instance of InboxService with the given database connection and context.
// Use this function to create a new InboxService instance, which can be used to interact with the inbox-related database tables.
func NewInboxService(db *gorm.DB, ctx *context.ERPContext) *InboxService {
	return &InboxService{db: db, ctx: ctx}
}

// GetInboxes retrieves a list of inbox models from the database for the given user ID and member ID.
// The function filters the inboxes by the given user ID and member ID. If the user ID and member ID are both nil,
// the function returns an empty list. If the user ID is not nil, the function returns the inboxes for the given user ID.
// If the member ID is not nil, the function returns the inboxes for the given member ID.
// If the user ID and member ID are not nil, the function returns the inboxes that match both the user ID and member ID.
// The function also creates a default inbox and a trash inbox if they do not exist, and returns them in the list.
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

// SendMessage sends an inbox message using the provided data.
//
// It first checks if the RecipientMemberID is provided and attempts to find the default inbox
// for the member. If the inbox is not found, it creates a new default inbox using the member's
// details. The function then assigns the inbox ID to the message data.
//
// Similarly, if the RecipientUserID is provided, it searches for the default inbox for the user.
// If it doesn't exist, it creates a new default inbox for the user and assigns the inbox ID to
// the message data.
//
// The function finally creates the message in the database. If any operation fails, it returns
// an error, otherwise, it returns nil upon successful message creation.

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

// GetMessageByInboxID retrieves a paginated list of messages for a given inbox ID.
//
// If the inbox ID is provided, it only returns messages that are not replies and are addressed
// to the inbox. If the inbox is associated with a member, it also returns messages that are
// replies to messages in the inbox and addressed to the member.
//
// The search parameter is optional and is used to filter the results by subject or message content.
// The function uses request parameters to modify the pagination and filtering behavior.
//
// The function returns a paginated page of InboxMessageModel objects and an error if the operation
// fails, allowing the caller to handle any database-related issues that might occur during the query
// process.
func (s *InboxService) GetMessageByInboxID(request http.Request, search string, inboxID *string) (paginate.Page, error) {

	pg := paginate.New()
	stmt := s.db.
		Preload("SenderUser").
		Preload("SenderMember.User").
		Preload("RecipientUser").
		Preload("RecipientMember.User")

	if inboxID != nil {
		var inbox models.InboxModel
		if err := s.db.Where("id = ?  ", inboxID).First(&inbox).Error; err != nil {
			return paginate.Page{}, err
		}
		stmt = stmt.Where("inbox_id = ? AND parent_inbox_message_id IS NULL", *inboxID)
		if inbox.MemberID != nil {
			// stmt = stmt.Or("parent_inbox_message_id IS NOT NULL AND recipient_member_id = ? and sender_member_id != ?", *inbox.MemberID, *inbox.MemberID)
		}
		// if inbox.UserID != nil {
		// 	stmt = stmt.Or("parent_inbox_message_id IS NOT NULL AND recipient_user_id = ? and sender_user_id != ?", *inbox.UserID, *inbox.UserID)
		// }
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

// GetDefaultInbox retrieves the default inbox for a given user or member.
//
// It accepts a user ID and a member ID as parameters, either of which may be nil.
// If the user ID is provided, the function searches for the default inbox associated
// with that user. If the member ID is provided, it searches for the default inbox
// associated with that member. The function returns the found InboxModel and a nil
// error if successful, or a nil InboxModel and an error if the inbox is not found
// or if a database error occurs during the query.
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

// GetSentMessages retrieves a paginated list of sent messages for the specified user or member.
//
// It accepts an HTTP request, a search string, and optional user and member IDs for filtering.
// The function returns a paginated page of InboxMessageModel objects and an error if the operation fails.
func (s *InboxService) GetSentMessages(request http.Request, search string, userID *string, memberID *string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.Where("parent_inbox_message_id IS NULL").
		Preload("SenderUser").
		Preload("SenderMember.User").
		Preload("RecipientUser").
		Preload("RecipientMember.User")
	if userID != nil && memberID != nil {
		stmt = stmt.Where("sender_user_id = ? AND sender_member_id = ?", userID, memberID)
	} else if userID != nil {
		stmt = stmt.Where("sender_user_id = ?", userID)
	} else if memberID != nil {
		stmt = stmt.Where(" sender_member_id = ?", memberID)
	}
	stmt = stmt.Joins("JOIN inbox ON inbox_messages.inbox_id = inbox.id").
		Where("inbox.is_trash = ?", false)
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

// CountUnreadSendMessage counts the unread sent messages for the specified user or member.
//
// It accepts optional user and member IDs and returns the count of unread sent messages and an error if the operation fails.
func (s *InboxService) CountUnreadSendMessage(userID *string, memberID *string) (int64, error) {
	var count int64
	if userID != nil && memberID != nil {
		if err := s.db.Model(&models.InboxMessageModel{}).
			Where("sender_user_id = ? AND sender_member_id = ?", userID, memberID).
			Where("read = ?", false).
			Where("parent_inbox_message_id IS NULL").
			Count(&count).Error; err != nil {
			return 0, err
		}
	} else if userID != nil {
		if err := s.db.Model(&models.InboxMessageModel{}).
			Where("sender_user_id = ? ", userID).
			Where("read = ?", false).
			Where("parent_inbox_message_id IS NULL").
			Count(&count).Error; err != nil {
			return 0, err
		}
	} else if memberID != nil {
		if err := s.db.Model(&models.InboxMessageModel{}).
			Where("sender_member_id = ? ", memberID).
			Where("read = ?", false).
			Where("parent_inbox_message_id IS NULL").
			Count(&count).Error; err != nil {
			return 0, err
		}
	}
	return count, nil
}

// CountUnread counts the unread messages for the specified user or member.
//
// It accepts optional user and member IDs and returns the count of unread messages and an error if the operation fails.
func (s *InboxService) CountUnread(userID *string, memberID *string) (int64, error) {
	var count int64
	if userID != nil {
		if err := s.db.Model(&models.InboxMessageModel{}).
			Joins("JOIN inbox ON inbox_messages.inbox_id = inbox.id").
			Where("inbox.is_trash = ?", false).
			Where("inbox.is_default = ?", true).
			Where("inbox.user_id = ?", *userID).
			Where("recipient_user_id = ?", *userID).
			Where("read = ?", false).
			Where("parent_inbox_message_id IS NULL").
			Count(&count).Error; err != nil {
			return 0, err
		}
	} else if memberID != nil {
		if err := s.db.Model(&models.InboxMessageModel{}).
			Joins("JOIN inbox ON inbox_messages.inbox_id = inbox.id").
			Where("inbox.is_trash = ?", false).
			Where("inbox.is_default = ?", true).
			Where("inbox.member_id = ?", *memberID).
			Where("recipient_member_id = ?", *memberID).
			Where("read = ?", false).
			Where("parent_inbox_message_id IS NULL").
			Count(&count).Error; err != nil {
			return 0, err
		}
	}
	return count, nil
}

// DeleteMessage moves the specified message to the trash for the given user or member.
//
// It accepts the message ID, and optional user and member IDs, and returns an error if the operation fails.
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

// GetInboxMessageDetail retrieves the details of a specific inbox message, including its replies.
//
// It accepts the message ID and returns the InboxMessageModel object with details and an error if the operation fails.
func (s *InboxService) GetInboxMessageDetail(inboxMessageID string) (*models.InboxMessageModel, error) {
	var inboxMessage models.InboxMessageModel
	if err := s.db.
		Preload("ParentInboxMessage", func(db *gorm.DB) *gorm.DB {
			return db.Preload("SenderUser").
				Preload("SenderMember.User").
				Preload("RecipientUser").
				Preload("RecipientMember.User")
		}).
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
