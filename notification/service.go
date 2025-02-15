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
	return &NotificationService{ctx: ctx}
}

func (s *NotificationService) Migrate() error {
	return s.ctx.DB.AutoMigrate(&models.NotificationModel{})
}

func (s *NotificationService) CreateNotification(notification *models.NotificationModel) error {
	return s.ctx.DB.Create(notification).Error
}

func (s *NotificationService) DeleteNotification(id string) error {
	return s.ctx.DB.Delete(&models.NotificationModel{}, "id = ?", id).Error
}

func (s *NotificationService) MarkAsRead(id string) error {
	return s.ctx.DB.Model(&models.NotificationModel{}).Where("id = ?", id).Update("is_read", true).Error
}

func (s *NotificationService) GetBrands(request http.Request, search string) (paginate.Page, error) {
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
	request.URL.Query().Get("page")
	stmt = stmt.Model(&models.NotificationModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.NotificationModel{})
	page.Page = page.Page + 1
	return page, nil
}
