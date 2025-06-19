package objects

type EmailData struct {
	Subject     string `json:"subject"`
	FullName    string `json:"full_name"`
	InvitedBy   string `json:"invited_by"`
	ProjectName string `json:"project_name"`
	CompanyName string `json:"company_name"`
	Email       string `json:"email"`
	Notif       string `json:"notif"`
	Link        string `json:"link"`
	Password    string `json:"password"`
}
