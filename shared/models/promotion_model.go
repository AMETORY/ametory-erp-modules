package models

import (
	"strings"
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PromotionModel struct {
	shared.BaseModel
	Name        string                 `gorm:"type:varchar(255);unique;not null" json:"name,omitempty"`
	Description string                 `json:"description,omitempty"`
	Type        string                 `gorm:"type:varchar(20);not null" json:"type,omitempty"` // discount, coupon, cashback, free_shipping
	StartDate   time.Time              `gorm:"not null" json:"start_date,omitempty"`
	EndDate     time.Time              `gorm:"not null" json:"end_date,omitempty"`
	IsActive    bool                   `gorm:"default:true" json:"is_active,omitempty"`
	Images      []FileModel            `gorm:"-" json:"images,omitempty"`
	Rules       []PromotionRuleModel   `gorm:"foreignKey:PromotionID;constraint:OnDelete:CASCADE" json:"rules,omitempty"`
	Actions     []PromotionActionModel `gorm:"foreignKey:PromotionID;constraint:OnDelete:CASCADE" json:"actions,omitempty"`
}

func (PromotionModel) TableName() string {
	return "promotions"
}

func (p *PromotionModel) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}

type PromotionRuleModel struct {
	shared.BaseModel
	PromotionID string                 `gorm:"type:char(36);not null;index" json:"promotion_id,omitempty"`
	Promotion   PromotionModel         `gorm:"foreignKey:PromotionID;constraint:OnDelete:CASCADE" json:"promotion,omitempty"`
	RuleType    string                 `gorm:"type:varchar(50);not null" json:"rule_type,omitempty"`
	RuleValue   string                 `gorm:"not null" json:"rule_value,omitempty"`
	Products    []ProductModel         `gorm:"-" json:"products,omitempty"`
	Categories  []ProductCategoryModel `gorm:"-" json:"categories,omitempty"`
}

func (PromotionRuleModel) TableName() string {
	return "promotion_rules"
}
func (p *PromotionRuleModel) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}

func (p *PromotionRuleModel) AfterFind(tx *gorm.DB) (err error) {
	if p.RuleType == "PRODUCTS" {
		err = tx.Model(&p.Products).Where("id IN (?)", strings.Split(p.RuleValue, ",")).Find(&p.Products).Error
	}
	if p.RuleType == "CATEGORIES" {
		err = tx.Model(&p.Categories).Where("id IN (?)", strings.Split(p.RuleValue, ",")).Preload("Products").Find(&p.Categories).Error
	}
	// Example: Process or modify data after fetching from the database
	// This is just a placeholder for any logic you might want to implement
	return
}

type PromotionActionModel struct {
	shared.BaseModel
	PromotionID string         `gorm:"type:char(36);not null;index" json:"promotion_id,omitempty"`
	Promotion   PromotionModel `gorm:"foreignKey:PromotionID;constraint:OnDelete:CASCADE" json:"promotion,omitempty"`
	ActionType  string         `gorm:"type:varchar(50);not null" json:"action_type,omitempty"` // discount, free_shipping, free_item
	ActionValue string         `gorm:"not null" json:"action_value,omitempty"`                 // Bisa angka (persentase atau nominal) atau item ID untuk free item
	// MinOrderQty int    `gorm:"default:0"`                 // Minimal jumlah item yang harus dibeli (untuk Buy 1 Get 1)
}

func (PromotionActionModel) TableName() string {
	return "promotion_actions"
}

func (p *PromotionActionModel) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}

func (b *PromotionModel) AfterFind(tx *gorm.DB) (err error) {
	var images []FileModel
	err = tx.Where("ref_id = ? and ref_type = ?", b.ID, "promotion").Find(&images).Error
	b.Images = images
	return err
}
