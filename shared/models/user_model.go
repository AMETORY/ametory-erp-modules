package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserModel adalah model database untuk user
type UserModel struct {
	shared.BaseModel
	FullName                   string             `gorm:"not null" json:"full_name,omitempty"`
	Username                   string             `gorm:"unique" json:"username,omitempty"`
	Code                       *string            `gorm:"unique;null" json:"code,omitempty"`
	Email                      string             `gorm:"unique;not null" json:"email,omitempty"`
	PhoneNumber                *string            `gorm:"null" json:"phone_number,omitempty"`
	Address                    string             `json:"address"`
	Password                   string             `gorm:"not null" json:"-"`
	VerifiedAt                 *time.Time         `gorm:"index" json:"verified_at,omitempty"`
	VerificationToken          string             `json:"verification_token,omitempty"`
	VerificationTokenExpiredAt *time.Time         `gorm:"index" json:"verification_token_expired_at,omitempty"`
	Roles                      []RoleModel        `gorm:"many2many:user_roles;constraint:OnDelete:CASCADE;" json:"roles,omitempty"`
	Companies                  []CompanyModel     `gorm:"many2many:user_companies;constraint:OnDelete:CASCADE;" json:"companies,omitempty"`
	Distributors               []DistributorModel `gorm:"many2many:user_distributors;constraint:OnDelete:CASCADE;" json:"distributors,omitempty"`
	ProfilePicture             *FileModel         `json:"profile_picture,omitempty" gorm:"-"`
	RoleID                     *string            `json:"role_id,omitempty" gorm:"-"`
	BirthDate                  *time.Time         `gorm:"null" json:"birth_date,omitempty"`
	Latitude                   float64            `json:"latitude" gorm:"type:decimal(10,8);default:-6.9951754;not null"`
	Longitude                  float64            `json:"longitude" gorm:"type:decimal(11,8);default:107.5946778;not null"`
	ProvinceID                 *string            `json:"province_id,omitempty" gorm:"type:char(2);index;constraint:OnDelete:SET NULL;"`
	RegencyID                  *string            `json:"regency_id,omitempty" gorm:"type:char(4);index;constraint:OnDelete:SET NULL;"`
	DistrictID                 *string            `json:"district_id,omitempty" gorm:"type:char(6);index;constraint:OnDelete:SET NULL;"`
	VillageID                  *string            `json:"village_id,omitempty" gorm:"type:char(10);index;constraint:OnDelete:SET NULL;"`
	IdentityNumber             string             `gorm:"type:varchar(255)" json:"identity_number,omitempty"`
	IdentityType               string             `gorm:"type:varchar(255)" json:"identity_type,omitempty"` // KTP, SIM, NPWP, dll
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

type PushTokenModel struct {
	shared.BaseModel
	Type       string  `json:"type"`
	DeviceType string  `json:"device_type"`
	Token      string  `json:"token" gorm:"uniqueIndex:idx_token"`
	UserID     *string `json:"user_id" gorm:"uniqueIndex:idx_token"`
	AdminID    *string `json:"admin_id" gorm:"uniqueIndex:idx_token"`
}

func (PushTokenModel) TableName() string {
	return "push_tokens"
}

func (pt *PushTokenModel) BeforeCreate(tx *gorm.DB) (err error) {
	if pt.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}
