package auth

import (
	"errors"
	"fmt"

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
func (s *AuthService) Register(username, email, password string) (*UserModel, error) {
	// Hash password
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	// Buat user baru
	user := UserModel{
		Username: username,
		Email:    email,
		Password: hashedPassword,
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

func (s *AuthService) Migrate() error {

	return s.db.AutoMigrate(&UserModel{})
}

func (s *AuthService) DB() *gorm.DB {
	return s.db
}
