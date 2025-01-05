package auth

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AdminModel struct {
	utils.BaseModel
	FullName                   string     `gorm:"not null"`
	Username                   string     `gorm:"unique"`
	Email                      string     `gorm:"unique;not null"`
	Password                   string     `gorm:"not null"`
	VerifiedAt                 *time.Time `gorm:"index"`
	VerificationToken          string
	VerificationTokenExpiredAt *time.Time  `gorm:"index"`
	Roles                      []RoleModel `gorm:"many2many:admin_roles;"`
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
