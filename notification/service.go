package notification

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
)

// NotificationService provides methods for managing notifications.
type NotificationService struct {
	ctx *context.ERPContext
}

// NewNotificationService creates a new instance of NotificationService.
// It initializes the service and performs migration if needed.
func NewNotificationService(ctx *context.ERPContext) *NotificationService {
	service := NotificationService{ctx: ctx}
	if !service.ctx.SkipMigration {
		service.Migrate()
	}
	return &service
}

// Migrate performs the database migration for the NotificationModel.
func (s *NotificationService) Migrate() error {
	return s.ctx.DB.AutoMigrate(&models.NotificationModel{})
}

// CreateNotification creates a new notification record in the database.
func (s *NotificationService) CreateNotification(notification *models.NotificationModel) error {
	return s.ctx.DB.Create(notification).Error
}

// CreateNotificationWithCallback creates a notification and executes a callback on success.
func (s *NotificationService) CreateNotificationWithCallback(notification *models.NotificationModel, callback func(notification *models.NotificationModel)) error {
	err := s.ctx.DB.Create(notification).Error
	if err != nil {
		return err
	}
	callback(notification)
	return nil
}

// DeleteNotification deletes a notification by ID.
func (s *NotificationService) DeleteNotification(id string) error {
	return s.ctx.DB.Delete(&models.NotificationModel{}, "id = ?", id).Error
}

// MarkAsRead marks a notification as read by ID.
func (s *NotificationService) MarkAsRead(id string) error {
	return s.ctx.DB.Model(&models.NotificationModel{}).Where("id = ?", id).Update("is_read", true).Error
}

// GetNotificationDetail retrieves notification details by ID.
func (s *NotificationService) GetNotificationDetail(id string) (*models.NotificationModel, error) {
	var notification models.NotificationModel
	err := s.ctx.DB.Where("id = ?", id).First(&notification).Error
	if err != nil {
		return nil, err
	}
	return &notification, nil
}

// CountNotificationUnread counts unread notifications for a user and company.
func (s *NotificationService) CountNotificationUnread(userID string, companyID string) (int64, error) {
	var count int64
	err := s.ctx.DB.Model(&models.NotificationModel{}).
		Where("user_id = ? AND is_read = false AND company_id = ?", userID, companyID).
		Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

// GetNotifications retrieves paginated notifications based on search criteria.
func (s *NotificationService) GetNotifications(request http.Request, search string, userID *string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.ctx.DB
	if search != "" {
		stmt = stmt.Where("description ILIKE ? OR title ILIKE ?", "%"+search+"%", "%"+search+"%")
	}
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	if request.Header.Get("ID-Merchant") != "" {
		stmt = stmt.Where("merchant_id = ?", request.Header.Get("ID-Merchant"))
	}
	if request.Header.Get("ID-Distributor") != "" {
		stmt = stmt.Where("distributor_id = ?", request.Header.Get("ID-Distributor"))
	}
	if userID != nil {
		stmt = stmt.Where("user_id = ?", userID)
	}
	if request.URL.Query().Get("order") != "" {
		stmt = stmt.Order(request.URL.Query().Get("order"))
	} else {
		stmt = stmt.Order("date DESC")
	}
	if request.URL.Query().Get("is_unread") != "" {
		stmt = stmt.Where("is_read = ?", request.URL.Query().Get("is_unread") != "1")
	}
	utils.FixRequest(&request)
	page := pg.With(stmt.Model(&models.NotificationModel{})).Request(request).Response(&[]models.NotificationModel{})
	page.Page = page.Page + 1
	return page, nil
}
