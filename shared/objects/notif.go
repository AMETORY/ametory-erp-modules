package objects

type NewUserData struct {
	FullName          string `json:"full_name"`
	Email             string `json:"email"`
	Password          string `json:"password"`
	VerificationToken string `json:"verification_token"`
}
