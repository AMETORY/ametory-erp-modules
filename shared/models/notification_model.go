package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NotificationModel struct {
	shared.BaseModel
	Title         string            `gorm:"type:varchar(255);not null" json:"title"`
	Description   string            `gorm:"type:text" json:"description"`
	RefType       string            `gorm:"type:varchar(255);index" json:"ref_type"`
	RefID         string            `gorm:"type:char(36);index" json:"ref_id"`
	UserID        *string           `gorm:"type:char(36);index" json:"user_id"`
	User          *UserModel        `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;" json:"user,omitempty"`
	MerchantID    *string           `gorm:"type:char(36);index" json:"merchant_id"`
	Merchant      *MerchantModel    `gorm:"foreignKey:MerchantID;constraint:OnDelete:CASCADE;" json:"merchant,omitempty"`
	DistributorID *string           `gorm:"type:char(36);index" json:"distributor_id"`
	Distributor   *DistributorModel `gorm:"foreignKey:DistributorID;constraint:OnDelete:CASCADE;" json:"distributor,omitempty"`
	CompanyID     *string           `gorm:"type:char(36);index" json:"company_id"`
	Company       *CompanyModel     `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE;" json:"company,omitempty"`
	IsRead        bool              `gorm:"type:boolean;default:false" json:"is_read"`
	Date          *time.Time        `json:"date"`
}

func (NotificationModel) TableName() string {
	return "notifications"
}

func (n *NotificationModel) BeforeCreate(tx *gorm.DB) (err error) {
	if n.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	tx.Statement.SetColumn("date", time.Now())
	return
}

func (n *NotificationModel) AfterFind(tx *gorm.DB) (err error) {
	if n.Date != nil {
		n.Date = &n.CreatedAt
		tx.Save(n)
	}
	return
}
