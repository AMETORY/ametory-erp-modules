package models

import (
	"time"

	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type VideoModel struct {
	shared.BaseModel
	Title       string                `gorm:"type:varchar(255);not null" json:"title"`
	Description string                `gorm:"type:text" json:"description"`
	Type        string                `gorm:"type:varchar(255);not null;default:ARTICLE" json:"type"` // 'Article type, e.g. news, blog, press release'
	Content     string                `gorm:"type:text" json:"content"`
	Slug        string                `gorm:"type:varchar(255);uniqueIndex" json:"slug,omitempty"`
	AuthorID    *string               `gorm:"type:char(36);index" json:"author_id,omitempty"`
	Author      *UserModel            `gorm:"foreignKey:AuthorID;constraint:OnDelete:CASCADE;" json:"author,omitempty"`
	PublishedAt *time.Time            `json:"published_at,omitempty"`
	CategoryID  *string               `gorm:"type:char(36)" json:"category_id,omitempty"`
	Category    *ContentCategoryModel `gorm:"-" json:"category,omitempty"`
	CompanyID   *string               `gorm:"type:char(36);index" json:"company_id,omitempty"`
	Company     *CompanyModel         `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE;" json:"company,omitempty"`
	Tags        []string              `gorm:"-" json:"tags,omitempty"`
	Thumbnail   []FileModel           `gorm:"-" json:"images,omitempty"`
	Files       []FileModel           `gorm:"-" json:"files,omitempty"`
	Status      string                `gorm:"type:varchar(255);default:DRAFT" json:"status,omitempty"` // 'DRAFT', 'PUBLISHED', 'REJECTED', 'DELETED'
}

func (VideoModel) TableName() string {
	return "videos"
}

func (v *VideoModel) BeforeCreate(tx *gorm.DB) error {
	if v.ID == "" {
		v.ID = uuid.New().String()
	}
	return nil
}
