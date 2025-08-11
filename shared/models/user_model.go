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
	FullName                   string             `gorm:"not null" json:"full_name,omitempty" bson:"fullName,omitempty"`
	Username                   string             `gorm:"unique" json:"username,omitempty" bson:"username,omitempty"`
	Code                       *string            `gorm:"unique;null" json:"code,omitempty" bson:"code,omitempty"`
	Email                      string             `gorm:"unique;not null" json:"email,omitempty" bson:"email,omitempty"`
	PhoneNumber                *string            `gorm:"null" json:"phone_number,omitempty" bson:"phoneNumber,omitempty"`
	Address                    string             `json:"address" bson:"address,omitempty"`
	Password                   string             `gorm:"not null" json:"-" bson:"-"`
	VerifiedAt                 *time.Time         `gorm:"index" json:"verified_at,omitempty" bson:"verifiedAt,omitempty"`
	VerificationToken          string             `json:"verification_token,omitempty" bson:"verificationToken,omitempty"`
	VerificationTokenExpiredAt *time.Time         `gorm:"index" json:"verification_token_expired_at,omitempty" bson:"verificationTokenExpiredAt,omitempty"`
	Roles                      []RoleModel        `gorm:"many2many:user_roles;constraint:OnDelete:CASCADE;" json:"roles,omitempty" bson:"roles,omitempty"`
	Role                       *RoleModel         `gorm:"-" json:"role,omitempty" bson:"role,omitempty"`
	Companies                  []CompanyModel     `gorm:"many2many:user_companies;constraint:OnDelete:CASCADE;-:migration" json:"companies,omitempty" bson:"companies,omitempty"`
	Distributors               []DistributorModel `gorm:"many2many:user_distributors;constraint:OnDelete:CASCADE;-:migration" json:"distributors,omitempty" bson:"distributors,omitempty"`
	ProfilePicture             *FileModel         `json:"profile_picture,omitempty" gorm:"-" bson:"profilePicture,omitempty"`
	RoleID                     *string            `json:"role_id,omitempty" gorm:"-" bson:"roleID,omitempty"`
	BirthDate                  *time.Time         `gorm:"null" json:"birth_date,omitempty" bson:"birthDate,omitempty"`
	Latitude                   float64            `json:"latitude" gorm:"type:decimal(10,8);" bson:"latitude,omitempty"`
	Longitude                  float64            `json:"longitude" gorm:"type:decimal(11,8);" bson:"longitude,omitempty"`
	ProvinceID                 *string            `json:"province_id,omitempty" gorm:"type:char(2);index;constraint:OnDelete:SET NULL;" bson:"provinceID,omitempty"`
	RegencyID                  *string            `json:"regency_id,omitempty" gorm:"type:char(4);index;constraint:OnDelete:SET NULL;" bson:"regencyID,omitempty"`
	DistrictID                 *string            `json:"district_id,omitempty" gorm:"type:char(6);index;constraint:OnDelete:SET NULL;" bson:"districtID,omitempty"`
	VillageID                  *string            `json:"village_id,omitempty" gorm:"type:char(10);index;constraint:OnDelete:SET NULL;" bson:"villageID,omitempty"`
	IdentityNumber             string             `gorm:"type:varchar(255)" json:"identity_number,omitempty" bson:"identityNumber,omitempty"`
	IdentityType               string             `gorm:"type:varchar(255)" json:"identity_type,omitempty" bson:"identityType,omitempty"`
	IsVerified                 bool               `json:"is_verified,omitempty" gorm:"-" bson:"isVerified,omitempty"`
	CustomerLevel              *string            `json:"customer_level,omitempty" bson:"customerLevel,omitempty"`
	QRCode                     *string            `json:"qrcode,omitempty" gorm:"-" bson:"qrcode,omitempty"`
	ReferralCode               *string            `json:"referral_code,omitempty" bson:"referralCode,omitempty"`
	Upline                     *UserModel         `gorm:"foreignKey:ReferralCode;references:Code" json:"upline,omitempty" bson:"upline,omitempty"`
	Downlines                  []UserModel        `gorm:"-" json:"downlines,omitempty" bson:"downlines,omitempty"`
	PhoneNumberVerifiedAt      *time.Time         `gorm:"index" json:"phone_number_verified_at,omitempty" bson:"phoneNumberVerifiedAt,omitempty"`
	Employee                   *EmployeeModel     `gorm:"-" json:"employee,omitempty" bson:"employee,omitempty"`
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
	EmployeeID *string `json:"employee_id" gorm:"uniqueIndex:idx_token"`
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

func (u *UserModel) AfterFind(tx *gorm.DB) error {
	u.IsVerified = u.VerifiedAt != nil

	file := FileModel{}
	err := tx.Where("ref_id = ? and ref_type = ?", u.ID, "user").Order("created_at desc").First(&file).Error
	if err == nil {
		u.ProfilePicture = &file
	}

	if len(u.Roles) == 1 {
		u.Role = &u.Roles[0]
	}
	return nil
}
