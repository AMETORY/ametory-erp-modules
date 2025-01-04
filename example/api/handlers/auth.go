package handlers

import (
	u "ametory-erp/utils"

	"github.com/AMETORY/ametory-erp-modules/auth"
	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	ctx *context.ERPContext
}

func NewAuthHandler(ctx *context.ERPContext) *AuthHandler {
	return &AuthHandler{ctx: ctx}
}

func (h *AuthHandler) LoginHandler(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"message": "Invalid input"})
		return
	}
	authSrv, ok := h.ctx.AuthService.(*auth.AuthService)
	if !ok {
		c.JSON(500, gin.H{"message": "Auth service is not available"})
		return
	}
	user, err := authSrv.Login(req.Email, req.Password)
	if err != nil {
		c.JSON(500, gin.H{"message": err.Error()})
		return
	}

	token, err := u.GenerateJWT(user.ID)
	if err != nil {
		c.JSON(500, gin.H{"message": "Failed to generate token"})
		return
	}
	c.JSON(200, gin.H{"token": token})
}
func (h *AuthHandler) RegisterHandler(c *gin.Context) {
	h.ctx.Request = c.Request

	// RegisterHandler is a handler function to create a new auth.
	authSrv, ok := h.ctx.AuthService.(*auth.AuthService)
	if !ok {
		c.JSON(500, gin.H{"message": "Auth service is not available"})
		return
	}
	input := struct {
		FullName string `json:"full_name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"message": "Invalid input"})
		return
	}

	if input.Password == "" {
		input.Password = utils.RandString(20)
	}
	username := utils.CreateUsernameFromFullName(input.FullName)
	newUser, err := authSrv.Register(input.FullName, username, input.Email, input.Password)
	if err != nil {
		c.JSON(500, gin.H{"message": "Failed to create auth"})
		return
	}

	sender := h.ctx.EmailSender
	if sender == nil {
		c.JSON(500, gin.H{"message": "Email sender is not available"})
		return
	}
	sender.SetAddress(newUser.FullName, newUser.Email)

	emailData := struct {
		Name     string
		Notif    string
		Link     string
		Password string
	}{
		Name:     newUser.FullName,
		Notif:    "Terima Kasih telah bergabung dengan Ametory",
		Link:     "https://ametory.com/verify?token=" + newUser.VerificationToken,
		Password: input.Password,
	}
	if err := sender.SendEmail("Welcome to Ametory", emailData, []string{}); err != nil {
		c.JSON(500, gin.H{"message": "Failed to send email"})
		return
	}
	c.JSON(200, gin.H{"message": "Auth created successfully"})
}
