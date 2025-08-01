package email_api

// EmailAPI is the interface for sending emails.
type EmailAPI interface {
	// SendEmail sends an email.
	//
	// It takes the following parameters:
	//  - from: The sender of the email.
	//  - domain: The domain of the email service.
	//  - apiKey: The API key of the email service.
	//  - subject: The subject of the email.
	//  - to: The recipient of the email.
	//  - message: The message of the email.
	//  - attachment: The attachment of the email.
	//
	// It returns an error if there is an issue with sending the email.
	SendEmail(from, domain, apiKey, apiSecret, subject, to, message string, attachment []string) error
}

// EmailApiService is the implementation of the EmailAPI interface.
type EmailApiService struct {
	From      string
	Domain    string
	ApiKey    string
	ApiSecret string
	Sender    EmailAPI
}

// NewEmailApiService creates a new instance of the EmailApiService.
func NewEmailApiService(from, domain, apiKey string, apiSecret string, sender EmailAPI) *EmailApiService {
	return &EmailApiService{
		From:      from,
		Domain:    domain,
		ApiKey:    apiKey,
		ApiSecret: apiSecret,
		Sender:    sender,
	}
}

// SendEmail sends an email via the EmailApiService.
func (s *EmailApiService) SendEmail(subject, to, message string, attachment []string) error {
	return s.Sender.SendEmail(s.From, s.Domain, s.ApiKey, s.ApiSecret, subject, to, message, attachment)
}
