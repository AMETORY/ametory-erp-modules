package models

import (
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GeminiHistoryModel struct {
	shared.BaseModel
	Input    string  `json:"input"`
	Output   string  `json:"output"`
	FileURL  string  `gorm:"type:varchar(255)" json:"file_url"`
	MimeType string  `gorm:"type:varchar(255)" json:"mime_type"`
	AgentID  *string `gorm:"type:char(36)" json:"agent_id"`
	IsModel  bool    `json:"is_model"`
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

type GeminiAgent struct {
	shared.BaseModel
	Name               string  `gorm:"type:varchar(255);not null" json:"name"`
	ApiKey             string  `gorm:"type:varchar(255);not null" json:"api_key"`
	Active             bool    `gorm:"default:true" json:"active"`
	SystemInstruction  string  `gorm:"type:text" json:"system_instruction"`
	Model              string  `gorm:"type:varchar(255);default:'gemini-1.5-flash'" json:"model"`
	SetTemperature     float32 `gorm:"type:float;default:1" json:"set_temperature"`
	SetTopK            int32   `gorm:"type:int;default:0" json:"set_top_k"`
	SetTopP            float32 `gorm:"type:float;default:0.92" json:"set_top_p"`
	SetMaxOutputTokens int32   `gorm:"type:int;default:256" json:"set_max_output_tokens"`
	ResponseMimetype   string  `gorm:"type:varchar(255);default:'application/json'" json:"response_mimetype"`
}

func (a *GeminiAgent) TableName() string {
	return "gemini_agents"
}

func (a *GeminiAgent) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}
	return nil
}
