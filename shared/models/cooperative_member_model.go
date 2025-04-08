package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CooperativeMemberModel struct {
	shared.BaseModel
	CompanyID              *string            `gorm:"size:36" json:"-" bson:"company_id,omitempty"`
	Company                CompanyModel       `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE;" json:"company,omitempty"`
	Name                   string             `json:"name"`
	MemberIDNumber         string             `json:"member_id_number" gorm:"member_id_number" sql:"member_id_number"`
	JoinDate               time.Time          `json:"join_date"`
	Active                 bool               `json:"active"`
	Email                  string             `json:"email"`
	Picture                *FileModel         `gorm:"-" bson:"picture,omitempty" json:"picture,omitempty"`
	PhoneNumber            string             `json:"phone_number"`
	Address                string             `json:"address"`
	City                   string             `json:"city"`
	ZipCode                string             `json:"zip_code"`
	Country                string             `json:"country"`
	ConnectedTo            *string            `gorm:"size:36" json:"connected_to" `
	User                   *UserModel         `gorm:"foreignKey:ConnectedTo;constraint:OnDelete:CASCADE;" json:"user,omitempty"`
	RoleID                 *string            `gorm:"type:char(36)" json:"role_id,omitempty"`
	Role                   *RoleModel         `json:"role,omitempty" gorm:"foreignKey:RoleID;constraint:OnDelete:CASCADE;"`
	TotalSavings           float64            `json:"total_savings" gorm:"-"`
	TotalLoans             float64            `json:"total_loans" gorm:"-"`
	TotalRemainLoans       float64            `json:"total_remain_loans" gorm:"-"`
	TotalTransactions      float64            `json:"total_transactions" gorm:"-"`
	TotalDisbursement      float64            `json:"total_disbursement" gorm:"-"`
	NetSurplusTransactions []TransactionModel `gorm:"-" bson:"net_surplus_transactions,omitempty" json:"net_surplus_transactions,omitempty"`
	ApprovedBy             *string            `gorm:"size:36" json:"approved_by" `
	ApprovedByUser         *UserModel         `gorm:"foreignKey:ApprovedBy;constraint:OnDelete:SET NULL;" json:"approved_by_user,omitempty"`
	ApprovedAt             *time.Time         `json:"approved_at"`
	Status                 string             `json:"status" gorm:"default:'PENDING'"`
}

func (CooperativeMemberModel) TableName() string {
	return "cooperative_members"
}

func (cm *CooperativeMemberModel) BeforeCreate(tx *gorm.DB) error {
	if cm.ID == "" {
		cm.ID = uuid.New().String()
	}
	var file FileModel
	err := tx.First(&file, "ref_id = ? AND ref_type = ?", cm.ID, "cooperative-member").Error
	if err == nil {
		cm.Picture = &file
	}
	return nil
}
