package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ProductFeedbackModel adalah model database untuk menampung data feedback dari user terhadap suatu product
type ProductFeedbackModel struct {
	shared.BaseModel
	Rating    uint8         `gorm:"not null" json:"rating"`
	Message   string        `gorm:"type:text;not null" json:"message"`
	ProductID string        `gorm:"type:char(36);not null" json:"product_id"`
	Product   ProductModel  `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE;" json:"product"`
	VariantID *string       `gorm:"type:char(36)" json:"variant_id"`
	Variant   *VariantModel `gorm:"foreignKey:VariantID;constraint:OnDelete:CASCADE;" json:"variant"`
	UserID    string        `gorm:"type:char(36);not null" json:"user_id"`
	User      UserModel     `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;" json:"user"`
	Status    string        `gorm:"type:varchar(20);default:PENDING" json:"status"`
	Date      *time.Time    `json:"date"`
}

func (ProductFeedbackModel) TableName() string {
	return "product_feedbacks"
}

func (n *ProductFeedbackModel) BeforeCreate(tx *gorm.DB) (err error) {
	if n.ID == "" {
		tx.Statement.SetColumn("id", uuid.New().String())
	}
	tx.Statement.SetColumn("date", time.Now())
	return
}
