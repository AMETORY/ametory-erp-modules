package auth

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserModel adalah model database untuk user
type UserModel struct {
	utils.BaseModel
	FullName                   string     `gorm:"not null"`
	Username                   string     `gorm:"unique"`
	Email                      string     `gorm:"unique;not null"`
	Password                   string     `gorm:"not null"`
	VerifiedAt                 *time.Time `gorm:"index"`
	VerificationToken          string
	VerificationTokenExpiredAt *time.Time `gorm:"index"`
}

func (u *UserModel) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}

func (UserModel) TableName() string {
	return "users"
}

// Migrate menjalankan migrasi database untuk model user
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&UserModel{})
}

// HashPassword mengenkripsi password menggunakan bcrypt
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// CheckPassword memverifikasi password dengan hash yang tersimpan
func CheckPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func (s *AuthService) Migrate() error {

	return s.db.AutoMigrate(&UserModel{})
}

func (s *AuthService) DB() *gorm.DB {
	return s.db
}
