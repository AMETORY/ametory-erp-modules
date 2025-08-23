package ai_generator

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"

	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/shared/objects"
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/morkid/paginate"
	"gorm.io/gorm"
)

type AiGeneratorService struct {
	ctx       *context.Context
	db        *gorm.DB
	factory   GeneratorFactory
	config    *GeneratorConfig
	Functions map[string]any
}

func NewAiGeneratorService(ctx *context.Context, db *gorm.DB, skipMigration bool) *AiGeneratorService {
	service := AiGeneratorService{
		ctx: ctx,
		db:  db,
		factory: func(config GeneratorConfig) (AiGenerator, error) {
			return nil, nil
		},
		config:    &GeneratorConfig{},
		Functions: make(map[string]any),
	}
	if !skipMigration {
		service.db.AutoMigrate(&models.AiAgentModel{}, &models.AiAgentHistory{})
	}
	return &service
}

func (e *AiGeneratorService) RegisterFunction(name string, fn any) {
	e.Functions[name] = fn
}
func (e *AiGeneratorService) RegisterFunctions(name []string, fn any) {
	for _, v := range name {
		e.Functions[v] = fn
	}
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

func (s *AiGeneratorService) ParseResponse(msg AiMessage, callback func(msg string, command string, params map[string]any)) error {
	var resp objects.AiResponse
	err := json.Unmarshal([]byte(msg.Content), &resp)
	if err != nil {
		return err
	}

	params, ok := resp.Params.(map[string]any)
	if ok {
		callback(resp.Response, resp.Command, params)
		return nil
	}
	callback(resp.Response, resp.Command, map[string]any{})

	return nil
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

func (s *AiGeneratorService) GenerateContentAndParseResponse(
	agent *models.AiAgentModel,
	generator *AiGenerator,
	systemInstruction,
	sender,
	redisKey,
	userMsg string,
	responseToUser func(sender, userMsg, response, redisKey string, totalTokenCount int32) error,
	addToHistory func(sender, userMsg, response, redisKey string) error,
	processCommandResp func(resp string),
	regenerateCommandResp bool,
) (*objects.AiResponse, error) {
	if generator == nil {
		return nil, fmt.Errorf(" generator is nil")
	}

	if agent == nil {
		return nil, fmt.Errorf(" agent is nil")
	}

	limit := agent.HistoryLength
	histories, err := s.GetHistories(&agent.ID, nil, &redisKey, nil, &limit)
	if err != nil {

		return nil, err
	}

	var his []AiMessage = []AiMessage{}
	isModel := true
	modelLimit := agent.HistoryLength
	modelHistories, _ := s.GetHistories(&agent.ID, nil, nil, &isModel, &modelLimit)

	for _, v := range modelHistories {
		his = append(his, AiMessage{
			Role:    "user",
			Content: v.Input,
		})

		his = append(his, AiMessage{
			Role:    "assistant",
			Content: v.Output,
		})
	}

	histories = ReverseHistories(histories)
	for _, v := range histories {

		his = append(his, AiMessage{
			Role:    "user",
			Content: v.Input,
		})

		his = append(his, AiMessage{
			Role:    "assistant",
			Content: v.Output,
		})

	}
	gen := *generator
	gen.SetSystemInstruction(systemInstruction)

	resp, err := gen.Generate(userMsg, nil, his)
	if err != nil {
		return nil, err
	}

	// fmt.Println("RESPONSE AI")
	// utils.LogJson(resp.Content)

	// err = s.ParseResponse(*resp, func(response string, command string, params map[string]any) {
	// 	fmt.Println("PARSED", response, command)
	// 	utils.LogJson(params)
	// })

	// if err != nil {
	// 	fmt.Println("ERROR PARSE RESPONSE", err, *resp)
	// }

	// PARSED RESPONSE
	var parsedResponse objects.AiResponse
	err = json.Unmarshal([]byte(resp.Content), &parsedResponse)
	if err != nil {
		return nil, err
	}

	parsedResponse.TotalTokenCount = resp.TotalTokenCount
	response := parsedResponse.Response
	err = responseToUser(sender, userMsg, response, redisKey, parsedResponse.TotalTokenCount)
	if err != nil {
		return nil, err
	}
	// sender, userMsg, response, redisKey string
	addToHistory(sender, userMsg, resp.Content, redisKey)
	fmt.Println("RESPONSE")
	utils.LogJson(parsedResponse)
	if parsedResponse.Type == "command" {
		resp, err := s.ProcessCommand(parsedResponse)
		if err == nil {
			// fmt.Println("PROCESS COMMAND")
			// utils.LogJson(resp)
			processCommandResp(resp)
			if regenerateCommandResp {
				return s.GenerateContentAndParseResponse(agent, generator, systemInstruction, sender, redisKey, resp, responseToUser, addToHistory, processCommandResp, false)
			}
		}

	}

	return &parsedResponse, nil

}

func (s *AiGeneratorService) ProcessCommand(parsedResponse objects.AiResponse) (string, error) {
	fmt.Println("call function", parsedResponse.Command)
	var params map[string]interface{} = parsedResponse.Params.(map[string]interface{})

	_, ok := s.Functions[parsedResponse.Command].(func(map[string]interface{}) (string, error))
	if ok {
		fmt.Println("with params", params)
		return s.Functions[parsedResponse.Command].(func(map[string]interface{}) (string, error))(params)
	} else {
		fmt.Println("REFLECT OF FN", reflect.TypeOf(s.Functions[parsedResponse.Command]))
	}
	return "", errors.New("command not found")
}

func ReverseHistories(histories []models.AiAgentHistory) []models.AiAgentHistory {
	result := make([]models.AiAgentHistory, len(histories))
	for i, h := range histories {
		result[len(histories)-i-1] = h
	}
	return result
}
