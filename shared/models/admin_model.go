package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AdminModel struct {
	shared.BaseModel
	FullName                   string      `gorm:"not null" json:"full_name,omitempty"`
	Username                   string      `gorm:"unique" json:"username,omitempty"`
	Email                      string      `gorm:"unique;not null" json:"email,omitempty"`
	Phone                      *string     `gorm:"null" json:"phone,omitempty"`
	Password                   *string     `gorm:"not null" json:"-"`
	VerifiedAt                 *time.Time  `gorm:"index" json:"verified_at,omitempty"`
	VerificationToken          string      `gorm:"index" json:"-"`
	VerificationTokenExpiredAt *time.Time  `gorm:"index" json:"verification_token_expired_at,omitempty"`
	Roles                      []RoleModel `gorm:"many2many:admin_roles;constraint:OnDelete:CASCADE;" json:"roles,omitempty"`
	ProfilePicture             *FileModel  `json:"profile_picture,omitempty" gorm:"-"`
	RoleID                     *string     `json:"role_id" gorm:"-"`
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

func (u *AdminModel) AfterFind(tx *gorm.DB) error {

	file := FileModel{}
	err := tx.Where("ref_id = ? and ref_type = ?", u.ID, "admin").Order("created_at desc").First(&file).Error
	if err == nil {
		u.ProfilePicture = &file
	}

	return nil
}
