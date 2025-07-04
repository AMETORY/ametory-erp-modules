package notification

import (
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
)

type NotificationService struct {
	ctx *context.ERPContext
}

func NewNotificationService(ctx *context.ERPContext) *NotificationService {
	service := NotificationService{ctx: ctx}
	if !service.ctx.SkipMigration {
		service.Migrate()
	}
	return &service
}

func (s *NotificationService) Migrate() error {
	return s.ctx.DB.AutoMigrate(&models.NotificationModel{})
}

func (s *NotificationService) CreateNotification(notification *models.NotificationModel) error {
	return s.ctx.DB.Create(notification).Error
}
func (s *NotificationService) CreateNotificationWithCallback(notification *models.NotificationModel, callback func(notification *models.NotificationModel)) error {
	err := s.ctx.DB.Create(notification).Error
	if err != nil {
		return err

	}
	callback(notification)
	return nil
}

func (s *NotificationService) DeleteNotification(id string) error {
	return s.ctx.DB.Delete(&models.NotificationModel{}, "id = ?", id).Error
}

func (s *NotificationService) MarkAsRead(id string) error {
	return s.ctx.DB.Model(&models.NotificationModel{}).Where("id = ?", id).Update("is_read", true).Error
}

func (s *NotificationService) GetNotificationDetail(id string) (*models.NotificationModel, error) {
	var notification models.NotificationModel
	err := s.ctx.DB.Where("id = ?", id).First(&notification).Error
	if err != nil {
		return nil, err
	}
	return &notification, nil
}

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
func (s *NotificationService) GetNotifications(request http.Request, search string, userID *string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.ctx.DB
	if search != "" {
		stmt = stmt.Where("description ILIKE ? OR title ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
		)
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
	request.URL.Query().Get("page")
	stmt = stmt.Model(&models.NotificationModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.NotificationModel{})
	page.Page = page.Page + 1
	return page, nil
}
