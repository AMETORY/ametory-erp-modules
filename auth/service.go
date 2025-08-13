package auth

import (
	"errors"
	"log"
	"time"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"gorm.io/gorm"
)

// AuthService is a service for managing user authentication.
//
// It provides methods for registering, logging in, changing passwords, and
// verifying tokens.
type AuthService struct {
	erpContext *context.ERPContext
	db         *gorm.DB
}

// NewAuthService creates a new instance of AuthService.
func NewAuthService(erpContext *context.ERPContext) *AuthService {
	var service = AuthService{erpContext: erpContext, db: erpContext.DB}
	if erpContext.SkipMigration {
		return &service
	}
	err := service.Migrate()
	if err != nil {
		log.Println("ERROR AUTH MIGRATE", err)
		panic(err)
	}
	return &service
}

// Migrate runs the database migrations.
func (s *AuthService) Migrate() error {
	return s.db.AutoMigrate(&models.UserModel{}, &models.RoleModel{}, &models.PermissionModel{}, &models.PushTokenModel{}, &models.UserActivityModel{})
}

// DB returns the underlying database connection.
func (s *AuthService) DB() *gorm.DB {
	return s.db
}

// Register creates a new user.
//
// It takes the full name, username, email, password, and phone number as
// arguments.
func (s *AuthService) Register(fullname, username, email, password, phoneNumber string) (*models.UserModel, error) {
	// Hash password
	hashedPassword, err := models.HashPassword(password)
	if err != nil {
		return nil, err
	}
	var verificationToken string
	var verificationTokenExpiredAt *time.Time
	// Generate verification token
	if email != "" {
		verificationToken = utils.RandString(32, false)
		exp := time.Now().AddDate(0, 0, 7)
		verificationTokenExpiredAt = &exp
	}

	// Buat user baru
	user := models.UserModel{
		FullName:                   fullname,
		Username:                   username,
		Email:                      email,
		Password:                   hashedPassword,
		VerificationToken:          verificationToken,
		VerificationTokenExpiredAt: verificationTokenExpiredAt,
		PhoneNumber:                &phoneNumber,
	}

	// Simpan ke database
	if err := s.db.Create(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

// CreateUser creates a new user in the database.
//
// It returns the created UserModel or an error if any occurs.
func (s *AuthService) CreateUser(user *models.UserModel) (*models.UserModel, error) {
	hashedPassword, err := models.HashPassword(user.Password)
	if err != nil {
		return nil, err
	}
	user.Password = hashedPassword
	if err := s.db.Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

// Login logs in a user.
//
// It takes the username or email and password as arguments.
func (s *AuthService) Login(usernameOrEmail, password string, checkVerified bool) (*models.UserModel, error) {
	var user models.UserModel

	// Cari user berdasarkan username atau email
	if err := s.db.Where("username = ? OR email = ? OR phone_number = ?", usernameOrEmail, usernameOrEmail, usernameOrEmail).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	// Periksa apakah user sudah terverifikasi
	if checkVerified {
		if user.VerifiedAt == nil {
			return nil, errors.New("user not verified")
		}
	}
	// Verifikasi password
	if err := models.CheckPassword(user.Password, password); err != nil {
		return nil, errors.New("invalid password")
	}

	return &user, nil
}

// ForgotPassword sends a password reset email.
//
// It takes the email as an argument.
func (s *AuthService) ForgotPassword(email string) error {
	var user models.UserModel

	// Cari user berdasarkan email
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return err
	}

	// Di sini Anda bisa mengirim email reset password
	// Contoh sederhana: print token reset password ke console
	resetToken := "reset-token-example" // Ganti dengan logika generate token yang sesungguhnya
	println("Reset token for", email, ":", resetToken)

	return nil
}

// ChangePassword changes the password for a user.
//
// It takes the old password and new password as arguments.
func (s *AuthService) ChangePassword(userID, oldPassword, newPassword string) error {
	var user models.UserModel

	// Cari user berdasarkan ID
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return err
	}

	// Verifikasi password lama
	if err := models.CheckPassword(user.Password, oldPassword); err != nil {
		return errors.New("invalid password")
	}

	// Hash password baru
	hashedPassword, err := models.HashPassword(newPassword)
	if err != nil {
		return err
	}

	user.Password = hashedPassword
	if err := s.db.Save(&user).Error; err != nil {
		return err
	}

	return nil
}

// Verification memverifikasi token reset password
func (s *AuthService) Verification(token, newPassword string) error {
	var user models.UserModel

	// Cari user berdasarkan token
	if err := s.db.Where("verification_token = ?", token).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("invalid token")
		}
		return err
	}

	// Verifikasi apakah token belum expired
	if time.Now().After(*user.VerificationTokenExpiredAt) {
		return errors.New("token has expired")
	}
	now := time.Now()

	user.VerificationToken = "" // Hapus token
	user.VerificationTokenExpiredAt = nil
	user.VerifiedAt = &now

	if err := s.db.Save(&user).Error; err != nil {
		return err
	}
	return nil
}

// GetUserDataByPhoneNumber retrieves a user by their phone number.
//
// It returns the UserModel if found, otherwise an error.
func (s *AuthService) GetUserDataByPhoneNumber(phoneNumber string) (*models.UserModel, error) {
	var user models.UserModel
	if err := s.db.Where("phone_number = ?", phoneNumber).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}
	return &user, nil
}

// GetUserByEmail checks if a user exists by their email.
//
// It returns true if found, otherwise false.
func (s *AuthService) GetUserByEmail(email string) bool {
	var user models.UserModel
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false
		}
	}
	return true
}

// GetUserByPhoneNumber checks if a user exists by their phone number.
//
// It returns true if found, otherwise false.
func (s *AuthService) GetUserByPhoneNumber(phone string) bool {
	var user models.UserModel
	if err := s.db.Where("phone_number = ?", phone).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false
		}
	}
	return true
}

// GetUserByEmailOrPhone retrieves a user by their email or phone number.
//
// It returns the UserModel if found, otherwise an error.
func (s *AuthService) GetUserByEmailOrPhone(emailOrPhone string) (*models.UserModel, error) {
	var user models.UserModel
	if err := s.db.Where("email = ? OR phone_number = ?", emailOrPhone, emailOrPhone).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// GetUserByID retrieves a user by their ID, preloading roles and permissions.
//
// It returns the UserModel with profile picture and permissions if found, otherwise an error.
func (s *AuthService) GetUserByID(userID string) (*models.UserModel, error) {
	var user models.UserModel
	if err := s.db.Preload("Roles", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "name", "is_super_admin").Preload("Permissions", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "name")
		})
	}).Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	file := models.FileModel{}
	s.db.Where("ref_id = ? and ref_type = ?", user.ID, "user").Order("updated_at desc").First(&file)
	if file.ID != "" {
		user.ProfilePicture = &file
	}
	for i, v := range user.Roles {
		if v.IsSuperAdmin {
			var Permissions []models.PermissionModel
			s.db.Find(&Permissions)
			user.Roles[i].Permissions = Permissions
		}
	}
	return &user, nil
}

// VerificationEmail memverifikasi token reset password
//
// It takes the token as an argument.
//
// It returns an error if the token is invalid or has expired.
func (s *AuthService) VerificationEmail(token string) error {
	var user models.UserModel

	// Cari user berdasarkan token
	if err := s.db.Where("verification_token = ?", token).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("invalid token")
		}
		return err
	}

	// Verifikasi apakah token belum expired
	if time.Now().After(*user.VerificationTokenExpiredAt) {
		return errors.New("token has expired")
	}

	// Tandai user sebagai verified
	now := time.Now()
	user.VerifiedAt = &now
	user.VerificationToken = ""
	user.VerificationTokenExpiredAt = nil

	if err := s.db.Save(&user).Error; err != nil {
		return err
	}
	return nil
}

// GetCompanies retrieves the companies of a user.
//
// It takes the user ID as an argument.
//
// It returns a slice of CompanyModel and an error.
func (s *AuthService) GetCompanies(userID string) ([]models.CompanyModel, error) {
	var user models.UserModel
	// fmt.Println("s.db.", s.db)
	// Cari user berdasarkan ID
	if err := s.db.Preload("Companies", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "name")
	}).Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	for i, v := range user.Companies {
		var merchants []models.MerchantModel
		s.db.Where("company_id = ?", v.ID).Find(&merchants)
		user.Companies[i].Merchants = merchants
	}

	return user.Companies, nil
}

// UpdateEmail updates the email of a user.
//
// It takes the user ID, email, verification token, and verification expired at as arguments.
//
// It returns an error if the user is already verified or if there is an error saving the user.
func (s *AuthService) UpdateEmail(userID string, email string, verificationToken string, verificationExpiredAt time.Time) error {
	user := &models.UserModel{}
	if err := s.db.Where("id = ?", userID).First(user).Error; err != nil {
		return err
	}
	if user.VerifiedAt != nil {
		return errors.New("user already verified")
	}

	user.Email = email
	user.VerificationToken = verificationToken
	user.VerificationTokenExpiredAt = &verificationExpiredAt

	if err := s.db.Save(user).Error; err != nil {
		return err
	}

	return nil
}

// UpdateReferralCode updates the referral code of a user.
//
// It takes the user ID and referral code as arguments.
//
// It returns an error if the user is already verified or if there is an error saving the user.
func (s *AuthService) UpdateReferralCode(userID string, referralCode string) error {
	user := &models.UserModel{}
	if err := s.db.Where("id = ?", userID).First(user).Error; err != nil {
		return err
	}
	if user.VerifiedAt != nil {
		return errors.New("user already verified")
	}

	if user.ReferralCode != nil {
		return errors.New("user already has a referral code")
	}

	existingUser := &models.UserModel{}
	err := s.db.Where("referral_code = ?", referralCode).First(existingUser).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("user with referral code not found")
	}

	user.ReferralCode = &referralCode

	if err := s.db.Save(user).Error; err != nil {
		return err
	}

	return nil
}

// UpdateAddress updates the address of a user.
//
// It takes the user ID and address as arguments.
//
// It returns an error if there is an error saving the user.
func (s *AuthService) UpdateAddress(userID string, address string) error {
	user := &models.UserModel{}
	if err := s.db.Where("id = ?", userID).First(user).Error; err != nil {
		return err
	}

	user.Address = address

	if err := s.db.Save(user).Error; err != nil {
		return err
	}

	return nil
}

// UpdateLatLng updates the latitude and longitude of a user.
//
// It takes the user ID, latitude, and longitude as arguments.
//
// It returns an error if there is an error saving the user.
func (s *AuthService) UpdateLatLng(userID string, latitude, longitude float64) error {
	user := &models.UserModel{}
	if err := s.db.Where("id = ?", userID).First(user).Error; err != nil {
		return err
	}

	user.Latitude = latitude
	user.Longitude = longitude

	if err := s.db.Save(user).Error; err != nil {
		return err
	}

	return nil
}

// CreatePushToken creates a push token for a user.
//
// It takes the user ID, token, device type, and token type as arguments.
//
// It returns a pointer to the created push token and an error.
func (s *AuthService) CreatePushToken(userID *string, token, deviceType, tokenType string) (*models.PushTokenModel, error) {
	// Check if the token already exists
	var existingPushToken models.PushTokenModel
	if err := s.db.Where("token = ? and user_id = ?", token, userID).First(&existingPushToken).Error; err == nil {
		return nil, errors.New("token already exists")
	}

	pushToken := &models.PushTokenModel{
		Token:      token,
		DeviceType: deviceType,
		UserID:     userID,
		Type:       tokenType,
	}

	if err := s.db.Create(pushToken).Error; err != nil {
		return nil, err
	}

	return pushToken, nil
}

// GetTokenFromUserID retrieves the push tokens of a user.
//
// It takes the user ID as an argument.
//
// It returns a slice of strings and an error.
func (s *AuthService) GetTokenFromUserID(userID string) ([]string, error) {
	var pushTokens []models.PushTokenModel
	if err := s.db.Where("user_id = ?", userID).Find(&pushTokens).Error; err != nil {
		return nil, err
	}
	tokens := make([]string, len(pushTokens))
	for i, token := range pushTokens {
		tokens[i] = token.Token
	}
	return tokens, nil
}
