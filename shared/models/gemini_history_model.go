package models

import (
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GeminiHistoryModel struct {
	shared.BaseModel
	Input    string
	Output   string
	FileURL  string `gorm:"type:varchar(255)" json:"file_url"`
	MimeType string `gorm:"type:varchar(255)" json:"mime_type"`
}

func (m *GeminiHistoryModel) TableName() string {
	return "gemini_histories"
}

func (m *GeminiHistoryModel) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	return nil
}
