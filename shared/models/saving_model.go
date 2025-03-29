package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SavingModel struct {
	shared.BaseModel
	CompanyID            *string                `gorm:"size:36" json:"company_id"`
	Company              *CompanyModel          `gorm:"foreignKey:CompanyID;references:ID" json:"company"`
	UserID               *string                `gorm:"size:36" json:"user_id"`
	User                 UserModel              `gorm:"foreignKey:UserID;references:ID" json:"user"`
	MemberID             *string                `gorm:"size:36" json:"member_id"`
	CooperativeMemberID  *string                `gorm:"size:36" json:"cooperative_member_id"`
	CooperativeMember    CooperativeMemberModel `gorm:"foreignKey:CooperativeMemberID;references:ID" json:"cooperative_member"`
	AccountDestinationID *string                `gorm:"size:36" json:"account_destination_id"`
	AccountDestination   AccountModel           `gorm:"foreignKey:AccountDestinationID;references:ID" json:"account_destination"`
	NetSurplusID         *string                `gorm:"size:36" json:"net_surplus_id"`
	SavingType           string                 `gorm:"type:varchar(20)" json:"saving_type"` //'Principal', 'Mandatory', 'Voluntary'
	Amount               float64                `json:"amount"`
	Notes                string                 `json:"notes" example:"notes"`
	Date                 *time.Time             `json:"date" bson:"date"`
}

func (s *SavingModel) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	return nil
}

func (s *SavingModel) TableName() string {
	return "savings"
}
