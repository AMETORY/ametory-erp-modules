package auth

import (
	"errors"
	"log"
	"time"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"gorm.io/gorm"
)

type AuthService struct {
	erpContext *context.ERPContext
	db         *gorm.DB
}

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

// Register membuat user baru
func (s *AuthService) Register(fullname, username, email, password, phoneNumber string) (*UserModel, error) {
	// Hash password
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	// Generate verification token
	verificationToken := utils.RandString(32, false)
	verificationTokenExpiredAt := time.Now().AddDate(0, 0, 7) // 7 hari

	// Buat user baru
	user := UserModel{
		FullName:                   fullname,
		Username:                   username,
		Email:                      email,
		Password:                   hashedPassword,
		VerificationToken:          verificationToken,
		VerificationTokenExpiredAt: &verificationTokenExpiredAt,
		PhoneNumber:                &phoneNumber,
	}

	// Simpan ke database
	if err := s.db.Create(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

// Login memverifikasi username/email dan password
func (s *AuthService) Login(usernameOrEmail, password string) (*UserModel, error) {
	var user UserModel

	// Cari user berdasarkan username atau email
	if err := s.db.Where("username = ? OR email = ? OR phone_number = ?", usernameOrEmail, usernameOrEmail, usernameOrEmail).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	// Periksa apakah user sudah terverifikasi
	if user.VerifiedAt == nil {
		return nil, errors.New("user not verified")
	}

	// Verifikasi password
	if err := CheckPassword(user.Password, password); err != nil {
		return nil, errors.New("invalid password")
	}

	return &user, nil
}

// ForgotPassword mengirim email reset password (contoh sederhana)
func (s *AuthService) ForgotPassword(email string) error {
	var user UserModel

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

// ChangePassword mengganti password
func (s *AuthService) ChangePassword(userID, oldPassword, newPassword string) error {
	var user UserModel

	// Cari user berdasarkan ID
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return err
	}

	// Verifikasi password lama
	if err := CheckPassword(user.Password, oldPassword); err != nil {
		return errors.New("invalid password")
	}

	// Hash password baru
	hashedPassword, err := HashPassword(newPassword)
	if err != nil {
		return err
	}

	// Ganti password di database
	user.Password = hashedPassword
	if err := s.db.Save(&user).Error; err != nil {
		return err
	}

	return nil
}

// Verification memverifikasi token reset password
func (s *AuthService) Verification(token, newPassword string) error {
	var user UserModel

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

func (s *AuthService) GetUserByPhoneNumber(phoneNumber string) bool {
	var user UserModel
	if err := s.db.Where("phone_number = ?", phoneNumber).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false
		}
	}
	return true
}
func (s *AuthService) GetUserByEmail(email string) bool {
	var user UserModel
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false
		}
	}
	return true
}
func (s *AuthService) GetUserByID(userID string) (*UserModel, error) {
	var user UserModel
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

	file := shared.FileModel{}
	s.db.Where("ref_id = ? and ref_type = ?", user.ID, "admin").First(&file)
	if file.ID != "" {
		user.ProfilePicture = &file
	}
	for i, v := range user.Roles {
		if v.IsSuperAdmin {
			var Permissions []PermissionModel
			s.db.Find(&Permissions)
			user.Roles[i].Permissions = Permissions
		}
	}
	return &user, nil
}

func (s *AuthService) VerificationEmail(token string) error {
	var user UserModel

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
