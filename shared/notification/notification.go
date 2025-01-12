package notification

import "gorm.io/gorm"

type NotificationProvider interface {
	SendNotification(to string, title string, message string, data interface{}, attachments []string) error
	SetTemplate(template string, layout string) error
}
type NotificationService struct {
	db       *gorm.DB
	provider NotificationProvider
}

func NewNotificationService(db *gorm.DB, provider NotificationProvider) *NotificationService {
	return &NotificationService{db: db, provider: provider}
}

func (n *NotificationService) SendNotification(to string, title string, message string, data interface{}, attachments []string) error {
	return n.provider.SendNotification(to, title, message, data, attachments)
}
