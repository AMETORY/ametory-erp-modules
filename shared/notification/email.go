package notification

import "github.com/AMETORY/ametory-erp-modules/thirdparty"

type EmailNotification struct {
	sender *thirdparty.SMTPSender
}

func NewEmailNotification(sender *thirdparty.SMTPSender) *EmailNotification {
	return &EmailNotification{
		sender: sender,
	}
}

func (e *EmailNotification) SendNotification(to string, title string, message string, data interface{}, attachments []string) error {
	e.sender.SetAddress(to, to)
	if data == nil {
		return e.sender.SendEmailWithTemplate(title, message, attachments)
	}
	return e.sender.SendEmail(title, data, attachments)
}

func (e *EmailNotification) SetTemplate(template string, layout string) *thirdparty.SMTPSender {
	return e.sender.SetTemplate(layout, template)
}
