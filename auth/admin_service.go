package auth

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type AdminAuthService struct {
	erpContext *context.ERPContext
	db         *gorm.DB
}

// NewAdminAuthService creates a new instance of AdminAuthService.
// It initializes the service with the provided ERPContext and performs
// database migrations unless the SkipMigration flag is set. If migration
// fails, it logs the error and panics.

func NewAdminAuthService(erpContext *context.ERPContext) *AdminAuthService {
	var service = AdminAuthService{erpContext: erpContext, db: erpContext.DB}
	if erpContext.SkipMigration {
		return &service
	}
	err := service.Migrate()
	if err != nil {
		log.Println("ERROR ADMIN AUTH MIGRATE", err)
		panic(err)
	}
	return &service
}

// Migrate runs the database migrations. It creates the tables for the
// admin models (AdminModel, RoleModel, and PermissionModel) if they do not
// exist yet.
func (s *AdminAuthService) Migrate() error {

	return s.db.AutoMigrate(&models.AdminModel{}, &models.RoleModel{}, &models.PermissionModel{})
}

// Register creates a new admin user.
//
// It takes the full name, username, email, password, and a boolean value indicating
// whether the user is being added or not as arguments. If isAdd is true, the user
// is immediately verified and the verification token is set to an empty string.
// If isAdd is false, a verification token is generated and the user must be
// verified using the Verification method.
//
// It returns the created admin user if successful, otherwise an error.

func (s *AdminAuthService) Register(fullname, username, email, password string, isAdd bool) (*models.AdminModel, error) {
	// Hash password
	hashedPassword, err := models.HashPassword(password)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	var verTokenExp = (time.Now().AddDate(0, 0, 7))
	var verificationAt, verificationTokenExpiredAt *time.Time
	// Generate verification token
	verificationToken := utils.RandString(32, false)
	verificationTokenExpiredAt = &verTokenExp // 7 hari
	if isAdd {
		verificationAt = &now
		verificationToken = ""
		verificationTokenExpiredAt = nil
	}

	// Buat user baru
	user := models.AdminModel{
		FullName:                   fullname,
		Username:                   username,
		Email:                      email,
		Password:                   &hashedPassword,
		VerificationToken:          verificationToken,
		VerificationTokenExpiredAt: verificationTokenExpiredAt,
		VerifiedAt:                 verificationAt,
	}

	// Simpan ke database
	if err := s.db.Create(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

// Login logs in a user.
//
// It takes the username or email and password as arguments.
//
// It returns the logged in user if successful, otherwise an error.
func (s *AdminAuthService) Login(usernameOrEmail, password string) (*models.AdminModel, error) {
	var user models.AdminModel

	// Cari user berdasarkan username atau email
	if err := s.db.Where("username = ? OR email = ?", usernameOrEmail, usernameOrEmail).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	// Periksa apakah user sudah terverifikasi
	if user.VerifiedAt == nil {
		return nil, errors.New("user not verified")
	}
	// fmt.Println("password", password)
	// Verifikasi password
	if err := models.CheckPassword(*user.Password, password); err != nil {
		return nil, errors.New("invalid password")
	}

	return &user, nil
}

// ForgotPassword sends a password reset email.
//
// It takes the email as an argument.
//
// It returns an error if the user is not found or if there is a problem with the database.
func (s *AdminAuthService) ForgotPassword(email string) error {
	var user models.AdminModel

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
// It takes the user ID, old password, and new password as arguments.
//
// It returns an error if the user is not found, the old password is invalid, or
// if there is a problem with the database.
func (s *AdminAuthService) ChangePassword(userID, oldPassword, newPassword string) error {
	var user models.AdminModel

	// Cari user berdasarkan ID
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return err
	}

	// Verifikasi password lama
	if err := models.CheckPassword(*user.Password, oldPassword); err != nil {
		return errors.New("invalid password")
	}

	// Hash password baru
	hashedPassword, err := models.HashPassword(newPassword)
	if err != nil {
		return err
	}

	// Ganti password di database
	user.Password = &hashedPassword
	if err := s.db.Save(&user).Error; err != nil {
		return err
	}

	return nil
}

// Verification verifies a user's reset password token and updates the password.
//
// It takes the token and new password as arguments. If the token is valid and
// not expired, the user's password is updated to the new password, and the token
// is invalidated. The user's verified status is also updated.
//
// Returns an error if the token is invalid, expired, or if there is an issue
// updating the user in the database.

func (s *AdminAuthService) Verification(token, newPassword string) error {
	var user models.AdminModel

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

// GetAdminByID retrieves a user by their ID, preloading roles and permissions.
//
// It returns the AdminModel with profile picture and permissions if found, otherwise an error.
func (s *AdminAuthService) GetAdminByID(userID string) (*models.AdminModel, error) {
	var user models.AdminModel
	// fmt.Println("s.db.", s.db)
	// Cari user berdasarkan ID
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
	s.db.Where("ref_id = ? and ref_type = ?", user.ID, "admin").First(&file)
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

// UpdateAdminByID updates the data of a user by their ID.
//
// It takes the user ID and data to be updated as arguments.
//
// It returns an error if the user is not found or if there is an error saving the user.
func (s *AdminAuthService) UpdateAdminByID(userID string, updatedData *models.AdminModel) error {
	var user models.AdminModel

	// Cari user berdasarkan ID
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return err
	}

	// Perbarui data user
	user.FullName = updatedData.FullName
	user.Phone = updatedData.Phone
	if err := s.db.Save(&user).Error; err != nil {
		return err
	}

	return nil
}

// GetAdmins retrieves a list of users with pagination.
//
// It takes a request and a search parameter as arguments.
//
// The search parameter is used to search the users by full name, email, or username.
//
// It returns a paginate.Page containing the list of AdminModel with profile picture.
func (s *AdminAuthService) GetAdmins(request http.Request, search string) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db.Preload("Roles")
	if search != "" {
		stmt = stmt.Where("admins.full_name ILIKE ? OR admins.email ILIKE ? OR admins.username ILIKE ?",
			"%"+search+"%",
			"%"+search+"%",
			"%"+search+"%",
		)
	}

	stmt = stmt.Model(&models.AdminModel{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.AdminModel{})
	page.Page = page.Page + 1
	items := page.Items.(*[]models.AdminModel)
	newItems := make([]models.AdminModel, 0)

	for _, v := range *items {
		img, err := s.GetProfilePicture(v.ID)
		if err == nil {
			v.ProfilePicture = &img
		}
		newItems = append(newItems, v)
	}
	page.Items = &newItems
	return page, nil
}

// GetProfilePicture returns the profile picture of a user by their ID.
//
// It takes the user ID as an argument.
//
// It returns the FileModel of the profile picture if found, otherwise an error.
func (s *AdminAuthService) GetProfilePicture(userID string) (models.FileModel, error) {
	var image models.FileModel
	err := s.db.Where("ref_id = ? and ref_type = ?", userID, "admin").First(&image).Error
	return image, err
}

// CreatePushToken creates a new push token for an admin
//
// It takes the admin ID, token, device type, and token type as arguments.
//
// It returns a pointer to the created push token and an error.
func (s *AdminAuthService) CreatePushToken(adminID *string, token, deviceType, tokenType string) (*models.PushTokenModel, error) {
	// Check if the token already exists
	var existingPushToken models.PushTokenModel
	if err := s.db.Where("token = ?  and admin_id = ?", token, adminID).First(&existingPushToken).Error; err == nil {
		return nil, errors.New("token already exists")
	}

	pushToken := &models.PushTokenModel{
		Token:      token,
		DeviceType: deviceType,
		Type:       tokenType,
		AdminID:    adminID,
	}

	if err := s.db.Create(pushToken).Error; err != nil {
		return nil, err
	}

	return pushToken, nil
}

// GetTokenFromAdminID returns a slice of strings containing the push tokens of an admin
//
// It takes the admin ID as an argument.
//
// It returns a slice of strings containing the push tokens if found, otherwise an error.
func (s *AuthService) GetTokenFromAdminID(userID string) ([]string, error) {
	var pushTokens []models.PushTokenModel
	if err := s.db.Where("admin_id = ?", userID).Find(&pushTokens).Error; err != nil {
		return nil, err
	}
	tokens := make([]string, len(pushTokens))
	for i, token := range pushTokens {
		tokens[i] = token.Token
	}
	return tokens, nil
}

// GetUserByEmail returns a user by their email
//
// It takes the email as an argument.
//
// It returns the UserModel if found, otherwise an error.
func (s *AdminAuthService) GetUserByEmail(email string) (*models.AdminModel, error) {
	var user models.AdminModel
	// Cari user berdasarkan email atau phone number
	if err := s.db.Where("email  = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// GetUserByID returns a user by their ID
//
// It takes the user ID as an argument.
//
// It returns the UserModel if found, otherwise an error.
func (s *AdminAuthService) GetUserByID(userID string) (*models.AdminModel, error) {
	var user models.AdminModel
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// VerificationEmail verifies a user by their verification token
//
// It takes the verification token as an argument.
//
// It returns an error if the token is invalid, expired, or if there is an issue updating the user in the database.
func (s *AdminAuthService) VerificationEmail(token string) error {
	var user models.AdminModel

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
