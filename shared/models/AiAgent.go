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
	Name                  string        `gorm:"type:varchar(255);not null" json:"name" bson:"name"`
	ApiKey                string        `gorm:"type:varchar(255);not null" json:"api_key" bson:"apiKey"`
	Active                bool          `gorm:"default:true" json:"active" bson:"active"`
	Host                  string        `gorm:"type:varchar(255)" json:"host" bson:"host"`
	SystemInstruction     string        `gorm:"type:text" json:"system_instruction" bson:"systemInstruction"`
	Model                 string        `gorm:"type:varchar(255);default:'gemini-1.5-flash'" json:"model" bson:"model"`
	ResponseMimetype      string        `gorm:"type:varchar(255);default:'application/json'" json:"response_mimetype" bson:"responseMimetype"`
	CompanyID             *string       `gorm:"type:char(36)" json:"company_id" bson:"company_id"`
	Company               *CompanyModel `gorm:"foreignKey:CompanyID;references:ID" json:"company" bson:"company"`
	AgentType             AiAgentType   `gorm:"type:varchar(255);default:'gemini'" json:"agent_type" bson:"agentType"`
	AutoResponseStartTime *string       `json:"auto_response_start_time" gorm:"type:varchar(32)"`
	AutoResponseEndTime   *string       `json:"auto_response_end_time" gorm:"type:varchar(32)"`
	NeedRegistration      bool          `json:"need_registration" bson:"needRegistration" default:"true"`
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
