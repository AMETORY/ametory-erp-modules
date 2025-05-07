package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MemberModel struct {
	shared.BaseModel
	CompanyID *string       `gorm:"uniqueIndex:idx_member;type:char(36)" json:"company_id,omitempty"`
	Company   *CompanyModel `json:"company,omitempty" gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE;"`
	UserID    string        `gorm:"uniqueIndex:idx_member;type:char(36)" json:"user_id,omitempty"`
	User      UserModel     `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;"`
	RoleID    *string       `gorm:"type:char(36)" json:"role_id,omitempty"`
	Role      *RoleModel    `json:"role,omitempty" gorm:"foreignKey:RoleID;constraint:OnDelete:CASCADE;"`
}

func (MemberModel) TableName() string {
	return "members"
}

func (m *MemberModel) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return nil
}

type MemberInvitationModel struct {
	shared.BaseModel
	CompanyID           *string       `gorm:"type:char(36)" json:"company_id,omitempty"`
	Company             *CompanyModel `json:"company,omitempty" gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE;"`
	UserID              string        `gorm:"type:char(36)" json:"user_id,omitempty"`
	User                UserModel     `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;"`
	TeamID              *string       `gorm:"type:char(36)" json:"team_id,omitempty"`
	Team                *TeamModel    `json:"team,omitempty" gorm:"foreignKey:TeamID;constraint:OnDelete:CASCADE;"`
	FullName            string        `gorm:"type:varchar(255)" json:"full_name"`
	ProjectID           *string       `gorm:"type:char(36)" json:"project_id,omitempty"`
	Project             *ProjectModel `json:"project,omitempty" gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE;"`
	RoleID              *string       `gorm:"type:char(36)" json:"role_id,omitempty"`
	Role                *RoleModel    `json:"role,omitempty" gorm:"foreignKey:RoleID;constraint:OnDelete:CASCADE;"`
	InviterID           string        `gorm:"type:char(36)" json:"inviter_id,omitempty"`
	Inviter             *UserModel    `gorm:"foreignKey:InviterID" json:"inviter,omitempty"`
	ExpiredAt           *time.Time    `json:"expired_at,omitempty"`
	Token               string        `gorm:"type:varchar(255)" json:"token,omitempty"`
	Email               string        `gorm:"type:varchar(255)" json:"email,omitempty"`
	IsCooperativeMember bool          `json:"is_cooperative_member"`
}

func (MemberInvitationModel) TableName() string {
	return "member_invitations"
}

func (m *MemberInvitationModel) BeforeCreate(tx *gorm.DB) (err error) {

	if m.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	m.Token = utils.RandString(32, false)
	if m.ExpiredAt == nil {
		expAt := time.Now().AddDate(0, 0, 7)
		m.ExpiredAt = &expAt
	}
	return nil
}
