package notification

type WhatsappNotification struct {
}

func NewWhatsappNotification() *WhatsappNotification {
	return &WhatsappNotification{}
}

func (w *WhatsappNotification) SendNotification(to string, title string, message string, data interface{}, attachments []string) error {
	// Implement the logic to send a WhatsApp notification here
	// if data is not nil, it should be a map[string]interface{} and parse from template
	return nil
}
func (w *WhatsappNotification) SetTemplate(template string, layout string) error {
	// Implement the logic to set the template for WhatsApp notification here
	return nil
}
