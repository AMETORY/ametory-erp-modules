package email_api

type EmailAPI interface {
	SendEmail(from, domain, apiKey, subject, to, message string, attachment []string) error
}

type EmailApiService struct {
	From   string
	Domain string
	ApiKey string
	Sender EmailAPI
}

func NewEmailApiService(from, domain, apiKey string, sender EmailAPI) *EmailApiService {
	return &EmailApiService{
		From:   from,
		Domain: domain,
		ApiKey: apiKey,
		Sender: sender,
	}
}
func (s *EmailApiService) SendEmail(subject, to, message string, attachment []string) error {
	return s.Sender.SendEmail(s.From, s.Domain, s.ApiKey, subject, to, message, attachment)
}
