package ai_generator

import (
	"context"
	"fmt"
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type AiGeneratorService struct {
	ctx     *context.Context
	db      *gorm.DB
	factory GeneratorFactory
	config  *GeneratorConfig
}

func NewAiGeneratorService(ctx *context.Context, db *gorm.DB, skipMigration bool) *AiGeneratorService {
	service := AiGeneratorService{
		ctx: ctx,
		db:  db,
		factory: func(config GeneratorConfig) (AiGenerator, error) {
			return nil, nil
		},
		config: &GeneratorConfig{},
	}
	if !skipMigration {
		service.db.AutoMigrate(&models.AiAgentModel{}, &models.AiAgentHistory{})
	}
	return &service
}

func (s *AiGeneratorService) SetFactory(factory GeneratorFactory) {
	s.factory = factory
}

func (s *AiGeneratorService) SetConfig(config GeneratorConfig) {
	s.config = &config
}

func (s *AiGeneratorService) CreateAgent(agent *models.AiAgentModel) error {
	err := s.db.Create(agent).Error
	if err != nil {
		return err
	}
	return nil
}

func (s *AiGeneratorService) GetAgent(id string) (*models.AiAgentModel, error) {
	agent := &models.AiAgentModel{}
	err := s.db.Where("id = ?", id).First(agent).Error
	if err != nil {
		return nil, err
	}
	return agent, nil
}

func (s *AiGeneratorService) GetGeneratorFromID(id string) (AiGenerator, error) {
	agent, err := s.GetAgent(id)
	if err != nil {
		return nil, fmt.Errorf("error fetching agent: %v", err)
	}
	var factory GeneratorFactory
	switch agent.AgentType {
	case models.AiAgentTypeDeepSeek:
		factory = func(config GeneratorConfig) (AiGenerator, error) {
			init := NewDeepSeekService(config.Ctx, config.APIKey)
			return init, nil
		}

	case models.AiAgentTypeGemini:
		factory = func(config GeneratorConfig) (AiGenerator, error) {
			init := NewGeminiV2Service(config.Ctx, config.APIKey)
			return init, nil
		}

	case models.AiAgentTypeOpenAI:
		factory = func(config GeneratorConfig) (AiGenerator, error) {
			init := NewOpenAiService(config.Ctx, config.APIKey)
			return init, nil
		}

	case models.AiAgentTypeOllama:
		factory = func(config GeneratorConfig) (AiGenerator, error) {
			init := NewOllamaService(config.Ctx, agent.Model)
			return init, nil
		}

	default:
		return nil, fmt.Errorf("unknown agent type: %s", agent.AgentType)
	}

	config := GeneratorConfig{
		Ctx:               s.ctx,
		APIKey:            agent.ApiKey,
		SystemInstruction: agent.SystemInstruction,
		Model:             agent.Model,
		Host:              agent.Host,
	}

	generator, err := factory(config)
	if err != nil {
		return nil, fmt.Errorf("error creating generator: %v", err)
	}

	generator.SetSystemInstruction(agent.SystemInstruction)
	var contentContentConfig ContentConfig = ContentConfig{
		ResponseMIMEType: "json",
	}
	generator.SetContentConfig(&contentContentConfig)
	generator.SetModel(agent.Model)
	return generator, nil
}

func (s *AiGeneratorService) UpdateAgent(agent *models.AiAgentModel) error {
	err := s.db.Save(agent).Error
	if err != nil {
		return err
	}
	return nil
}

func (s *AiGeneratorService) DeleteAgent(id string) error {
	err := s.db.Where("id = ?", id).Delete(&models.AiAgentModel{}).Error
	if err != nil {
		return err
	}
	return nil
}

func (s *AiGeneratorService) GetAgents(request *http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.db
	if request.Header.Get("ID-Company") != "" {
		stmt = stmt.Where("company_id = ?", request.Header.Get("ID-Company"))
	}
	stmt = stmt.Order("created_at desc").Model(&models.AiAgentModel{})
	utils.FixRequest(request)
	page := pg.With(stmt).Request(request).Response(&[]models.AiAgentModel{})
	page.Page = page.Page + 1
	return page, nil
}

func (s *AiGeneratorService) CreateHistory(history *models.AiAgentHistory) error {
	err := s.db.Create(history).Error
	if err != nil {
		return err
	}
	return nil
}

func (s *AiGeneratorService) UpdateHistory(history *models.AiAgentHistory) error {
	err := s.db.Save(history).Error
	if err != nil {
		return err
	}
	return nil
}
func (s *AiGeneratorService) GetHistories(id, companyID *string, sessionCode *string, isModel *bool, limit *int) ([]models.AiAgentHistory, error) {
	stmt := s.db.Model(&models.AiAgentHistory{}).Where("ai_agent_id = ?", *id)
	if sessionCode != nil {
		stmt = stmt.Where("session_code = ?", *sessionCode)
	} else {
		stmt = stmt.Where("session_code IS NULL")
	}

	if companyID != nil {
		stmt = stmt.Where("company_id = ?", *companyID)
	}

	if isModel != nil {
		stmt = stmt.Where("is_model = ?", *isModel)
	}

	if limit != nil {
		stmt = stmt.Limit(*limit)
	}

	var histories []models.AiAgentHistory
	err := stmt.Order("created_at DESC").Find(&histories).Error
	if err != nil {
		return nil, err
	}
	return histories, nil
}

func (s *AiGeneratorService) DeleteHistory(id string) error {
	err := s.db.Where("id = ?", id).Delete(&models.AiAgentHistory{}).Error
	if err != nil {
		return err
	}
	return nil
}
