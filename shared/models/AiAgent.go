package models

import (
	"github.com/AMETORY/ametory-erp-modules/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AiAgentType string

const (
	AiAgentTypeGemini   AiAgentType = "gemini"
	AiAgentTypeOpenAI   AiAgentType = "openai"
	AiAgentTypeDeepSeek AiAgentType = "deepseek"
	AiAgentTypeOllama   AiAgentType = "ollama"
)

type AiAgentModel struct {
	shared.BaseModel
	Name              string        `gorm:"type:varchar(255);not null" json:"name"`
	ApiKey            string        `gorm:"type:varchar(255);not null" json:"api_key"`
	Active            bool          `gorm:"default:true" json:"active"`
	Host              string        `gorm:"type:varchar(255)" json:"host"`
	SystemInstruction string        `gorm:"type:text" json:"system_instruction"`
	Model             string        `gorm:"type:varchar(255);default:'gemini-1.5-flash'" json:"model"`
	ResponseMimetype  string        `gorm:"type:varchar(255);default:'application/json'" json:"response_mimetype"`
	CompanyID         *string       `gorm:"type:char(36)" json:"company_id"`
	Company           *CompanyModel `gorm:"foreignKey:CompanyID;references:ID"`
	AgentType         AiAgentType   `gorm:"type:varchar(255);default:'gemini'" json:"agent_type"`
}

func (AiAgentModel) TableName() string {
	return "ai_agents"
}

func (m *AiAgentModel) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	return nil
}

type AiAgentHistory struct {
	shared.BaseModel
	Input       string        `json:"input"`
	Output      string        `json:"output"`
	FileURL     string        `gorm:"type:varchar(255)" json:"file_url"`
	MimeType    string        `gorm:"type:varchar(255)" json:"mime_type"`
	AiAgentID   *string       `gorm:"type:char(36)" json:"ai_agent_id"`
	AiAgent     *AiAgentModel `gorm:"foreignKey:AiAgentID;references:ID"`
	CompanyID   *string       `gorm:"type:char(36)" json:"company_id"`
	Company     *CompanyModel `gorm:"foreignKey:CompanyID;references:ID"`
	Role        string        `json:"role" gorm:"type:varchar(255);default:'model'"`
	SessionCode *string       `gorm:"type:varchar(255)" json:"session_code"`
	IsModel     bool          `json:"is_model" gorm:"default:false"`
}

func (AiAgentHistory) TableName() string {
	return "ai_agent_histories"
}

func (m *AiAgentHistory) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	return nil
}
