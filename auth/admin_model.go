package auth

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AdminModel struct {
	utils.BaseModel
	FullName                   string            `gorm:"not null" json:"full_name"`
	Username                   string            `gorm:"unique" json:"username"`
	Email                      string            `gorm:"unique;not null" json:"email"`
	Phone                      *string           `gorm:"null" json:"phone"`
	Password                   *string           `gorm:"not null" json:"-"`
	VerifiedAt                 *time.Time        `gorm:"index" json:"verified_at"`
	VerificationToken          string            `gorm:"index" json:"-"`
	VerificationTokenExpiredAt *time.Time        `gorm:"index" json:"-"`
	Roles                      []RoleModel       `gorm:"many2many:admin_roles;" json:"-"`
	ProfilePicture             *shared.FileModel `json:"profile_picture" gorm:"-"`
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
