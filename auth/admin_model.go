package auth

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AdminModel struct {
	shared.BaseModel
	FullName                   string            `gorm:"not null" json:"full_name,omitempty"`
	Username                   string            `gorm:"unique" json:"username,omitempty"`
	Email                      string            `gorm:"unique;not null" json:"email,omitempty"`
	Phone                      *string           `gorm:"null" json:"phone,omitempty"`
	Password                   *string           `gorm:"not null" json:"-"`
	VerifiedAt                 *time.Time        `gorm:"index" json:"verified_at,omitempty"`
	VerificationToken          string            `gorm:"index" json:"-"`
	VerificationTokenExpiredAt *time.Time        `gorm:"index" json:"verification_token_expired_at,omitempty"`
	Roles                      []RoleModel       `gorm:"many2many:admin_roles;" json:"roles,omitempty"`
	ProfilePicture             *shared.FileModel `json:"profile_picture" gorm:"-"`
	RoleID                     *string           `json:"role_id" gorm:"-"`
}

func (u *AdminModel) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return
}

func (AdminModel) TableName() string {
	return "admins"
}

func (s *AdminAuthService) Migrate() error {

	return s.db.AutoMigrate(&AdminModel{}, &RoleModel{}, &PermissionModel{})
}
