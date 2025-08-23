package chat_bot

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/AMETORY/ametory-erp-modules/shared/objects"
	"github.com/AMETORY/ametory-erp-modules/thirdparty/ai_generator"
	"github.com/AMETORY/ametory-erp-modules/thirdparty/redis"
	"github.com/AMETORY/ametory-erp-modules/utils"
)

type ChatBotService struct {
	ctx          *context.ERPContext
	redisService *redis.RedisService
	aiGenerator  *ai_generator.AiGeneratorService
	Functions    map[string]any
}

func NewChatBotService(ctx *context.ERPContext, redisService *redis.RedisService, aiGenerator *ai_generator.AiGeneratorService) *ChatBotService {
	return &ChatBotService{ctx: ctx,
		redisService: redisService,
		aiGenerator:  aiGenerator,
		Functions:    make(map[string]any),
	}
}

func (e *ChatBotService) RegisterFunction(name string, fn any) {
	e.Functions[name] = fn
}
func (e *ChatBotService) RegisterFunctions(names []string, fn any) {
	for _, name := range names {
		e.Functions[name] = fn
	}
}

func (e *ChatBotService) RunChatBot(
	body objects.MsgObject,
	redisKey,
	userMsg string,
	chatbotFlow *models.ChatbotFlow,
	responseToUser func(sender, userMsg, response, redisKey string, totalTokenCount int32),
	generateAiContent func(generator *ai_generator.AiGenerator, agent *models.AiAgentModel, sender, redisKey, userMsg string) (*objects.AiResponse, error),
) error {

	if chatbotFlow == nil {
		return errors.New("chatbot flow not found")
	}

	if e.aiGenerator == nil {
		return errors.New("redis service not initialized")
	}
	if e.redisService == nil {
		return errors.New("ai generator service not initialized")
	}

	var lastStep = 0
	var lastStatus = ""
	var lastFlow *models.Flow
	var collectedData = make(map[string]any)
	redisKeyState := redisKey + ":state"
	fmt.Println("redisKeyState", redisKeyState)

	// GET KEYWORDS
	keyword := strings.ToLower(strings.TrimSpace(userMsg))
	var foundFlow string
	var found bool
	for k, v := range chatbotFlow.Keywords {
		if strings.ToLower(strings.TrimSpace(k)) == keyword {
			foundFlow = v
			keyword = k
			found = true
		}
	}

	// GET LAST STATE
	lastStateStr, err := e.redisService.Client.LRange(*e.ctx.Ctx, redisKeyState, -1, -1).Result()
	if err == nil && len(lastStateStr) > 0 {
		var lastState map[string]any
		err = json.Unmarshal([]byte(lastStateStr[0]), &lastState)
		if err == nil {
			foundFlow = lastState["key"].(string)
			lst, ok := lastState["step"].(float64)
			if ok {
				lastStep = int(lst)
			}
			stts, ok := lastState["status"].(string)
			if ok {
				lastStatus = stts
			}
			colDt, ok := lastState["collected_data"].(map[string]any)
			if ok {
				collectedData = colDt
			} else {
				fmt.Println("ERROR GET COLLECTED DATA")
			}

			fmt.Println("LAST COLLECTED DATA")
			utils.LogJson(collectedData)
			// utils.LogJson(lastState)
			for k, v := range chatbotFlow.Flows {
				if k == lastState["key"].(string) {
					lastFlow = &v
				}
			}
		}
	}

	// CHECK LAST USER FLOW
	if lastFlow != nil {
		if lastFlow.Type == "menu" {
			for _, option := range lastFlow.Options {
				if option.Input == keyword {
					if option.NextFlow == "" && option.Response != "" {
						responseToUser(body.Sender, userMsg, option.Response, redisKey, 0)
						e.saveRedis(redisKeyState, map[string]any{
							"flow":      lastFlow,
							"key":       foundFlow,
							"userInput": userMsg,
							"step":      0,
							"status":    "",
						})
						return nil
					} else if option.NextFlow != "" {
						foundFlow = option.NextFlow
						found = true
					}
				}
			}
		}
		if lastFlow.Type == "form" {
			// foundFlow = lastFlow
		}

	}

	flow, ok := chatbotFlow.Flows[foundFlow]
	if ok && found {
		if flow.Type == "agent" && flow.Agent != nil && flow.AgentID != nil {

			generator, err := e.aiGenerator.GetGeneratorFromID(*flow.AgentID)
			if err != nil {
				return err
			}
			var agent models.AiAgentModel
			e.ctx.DB.Where("id = ?", *flow.AgentID).First(&agent)
			flow.Agent = &agent
			_, err = generateAiContent(&generator, flow.Agent, body.Sender, redisKey, userMsg)
			if err != nil {
				return err
			}
			e.saveRedis(redisKeyState, map[string]any{
				"flow":      flow,
				"key":       foundFlow,
				"userInput": userMsg,
				"step":      0,
				"status":    "",
			})
			return nil
		}

		if flow.Type == "menu" {
			responseToUser(body.Sender, userMsg, e.renderMenu(flow), redisKey, 0)
			e.saveRedis(redisKeyState, map[string]any{
				"flow":      flow,
				"key":       foundFlow,
				"userInput": userMsg,
				"step":      0,
				"status":    "",
			})
		}

		if flow.Type == "form" {
			// var menus []string
			if len(flow.Steps) > lastStep {
				currentStep := flow.Steps[lastStep]
				if lastStatus == "waiting_input" {
					fmt.Println("CURRENT STEP")
					utils.LogJson(currentStep)
					if currentStep.Validation.Type != "" {
						isFailed := false
						errorMessage := ""
						if currentStep.Validation.Type == "min_length" {
							minLength, err := strconv.Atoi(currentStep.Validation.Value)
							if err == nil {
								if len(userMsg) < minLength {
									isFailed = true
									errorMessage = currentStep.Validation.ErrorMessage
								}
							}
						}
						if currentStep.Validation.Type == "email" {
							fmt.Println("VALIDATION")
							utils.LogJson(currentStep.Validation)
							if !utils.IsValidEmail(userMsg) {
								isFailed = true
								errorMessage = currentStep.Validation.ErrorMessage
							}
						}
						if currentStep.Validation.Type == "regex" {
							if !regexp.MustCompile(currentStep.Validation.Pattern).MatchString(userMsg) {
								isFailed = true
								errorMessage = currentStep.Validation.ErrorMessage
							}
						}

						if isFailed {
							responseToUser(body.Sender, userMsg, errorMessage, redisKey, 0)
							e.saveRedis(redisKeyState, map[string]any{
								"flow":      flow,
								"key":       foundFlow,
								"userInput": userMsg,
								"step":      lastStep,
								"status":    "waiting_input",
							})
							return nil
						}

					}
					// SAVE TO DATA
					collectedData[currentStep.Field] = userMsg
					e.saveRedis(redisKeyState, map[string]any{
						"flow":           flow,
						"key":            foundFlow,
						"userInput":      userMsg,
						"step":           lastStep,
						"status":         "",
						"collected_data": collectedData,
					})

					fmt.Println("CURRENT COLLECTED DATA")
					utils.LogJson(collectedData)

					// fmt.Println("TOTAL STEP", len(flow.Steps))
					// fmt.Println("TOTAL STEP #2", lastStep+1)

					if len(flow.Steps) == lastStep+1 {
						fmt.Println("FORM COMPLETED", foundFlow)
						utils.LogJson(collectedData)

						if e.Functions[foundFlow] != nil {
							e.Functions[foundFlow].(func(data map[string]any))(collectedData)
						}
						// SAVE FORM DATA

						responseToUser(body.Sender, userMsg, flow.CompletionMessage, redisKey, 0)
						if flow.BackToFlow != "" {
							foundFlow = flow.BackToFlow
							lastStep = 0
							nextFlow, ok := chatbotFlow.Flows[foundFlow]
							if ok {
								flow = nextFlow
							}

							if flow.Type == "menu" {
								responseToUser(body.Sender, userMsg, e.renderMenu(flow), redisKey, 0)
								e.saveRedis(redisKeyState, map[string]any{
									"flow":      flow,
									"key":       foundFlow,
									"userInput": userMsg,
									"step":      0,
									"status":    "",
								})
								return nil
							}
						}
					} else {
						lastStep++
						currentStep = flow.Steps[lastStep]
					}

				}

				// NEXT STEP
				// fmt.Println("NEXT STEP")
				// utils.LogJson(currentStep)
				responseToUser(body.Sender, userMsg, currentStep.Question, redisKey, 0)
				e.saveRedis(redisKeyState, map[string]any{
					"flow":           flow,
					"key":            foundFlow,
					"userInput":      userMsg,
					"step":           lastStep,
					"status":         "waiting_input",
					"collected_data": collectedData,
				})
			}

		}
	} else {

		if chatbotFlow.FallbackResponseType == "text" {
			responseToUser(body.Sender, userMsg, chatbotFlow.FallbackResponse, redisKey, 0)
			return nil
		}
		if chatbotFlow.FallbackResponseType == "flow" {
			var fallbackFlow = chatbotFlow.Flows[chatbotFlow.FallbackResponse]

			if fallbackFlow.Type == "agent" && fallbackFlow.Agent != nil {
				generator, err := e.aiGenerator.GetGeneratorFromID(*fallbackFlow.AgentID)
				if err != nil {
					return err
				}
				var agent models.AiAgentModel
				e.ctx.DB.Where("id = ?", *fallbackFlow.AgentID).First(&agent)
				fmt.Println("FALLBACK FLOW", fallbackFlow.Type, fallbackFlow.Agent.SystemInstruction)
				utils.LogJson(fallbackFlow.Agent)
				fallbackFlow.Agent = &agent
				_, err = generateAiContent(&generator, fallbackFlow.Agent, body.Sender, redisKey, userMsg)
				if err != nil {
					return err
				}

				e.saveRedis(redisKeyState, map[string]any{
					"flow":           flow,
					"key":            foundFlow,
					"userInput":      userMsg,
					"step":           0,
					"status":         "",
					"collected_data": make(map[string]any),
				})
				return nil
			}

		}
	}

	return nil
}

func (e *ChatBotService) renderMenu(flow models.Flow) string {

	var menus []string
	for _, option := range flow.Options {
		menus = append(menus, fmt.Sprintf("[%s] %s", option.Input, option.Display))
	}
	respToUser := fmt.Sprintf(`%s

%s`, flow.Text, strings.Join(menus, "\n\n"))

	return respToUser
}

func (e *ChatBotService) saveRedis(key string, data map[string]any) {
	b, _ := json.Marshal(data)
	e.redisService.Client.RPush(*e.ctx.Ctx, key, string(b))
}
