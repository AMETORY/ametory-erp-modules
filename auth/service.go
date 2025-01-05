package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/AMETORY/ametory-erp-modules/utils"
	"gorm.io/gorm"
)

type AuthService struct {
	db *gorm.DB
}

func NewAuthService(db *gorm.DB) *AuthService {
	var service = AuthService{db: db}
	err := service.Migrate()
	if err != nil {
		fmt.Println("Error migrating auth service", err)
		return nil
	}
	return &service
}

// Register membuat user baru
func (s *AuthService) Register(fullname, username, email, password string) (*UserModel, error) {
	// Hash password
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	// Generate verification token
	verificationToken := utils.RandString(32)
	verificationTokenExpiredAt := time.Now().AddDate(0, 0, 7) // 7 hari

	// Buat user baru
	user := UserModel{
		FullName:                   fullname,
		Username:                   username,
		Email:                      email,
		Password:                   hashedPassword,
		VerificationToken:          verificationToken,
		VerificationTokenExpiredAt: &verificationTokenExpiredAt,
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
