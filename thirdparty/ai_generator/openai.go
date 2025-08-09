package ai_generator

import (
	"context"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/shared"
)

type OpenAiService struct {
	ctx               *context.Context
	ApiKey            string
	client            *openai.Client
	systemInstruction string
	model             string
}

func NewOpenAiService(ctx *context.Context, apiKey string) *OpenAiService {
	client := openai.NewClient(
		option.WithAPIKey(apiKey), // defaults to os.LookupEnv("OPENAI_API_KEY")
	)

	return &OpenAiService{
		ctx:    ctx,
		ApiKey: apiKey,
		client: &client,
		model:  openai.ChatModelGPT4o,
	}
}

func (g *OpenAiService) Generate(prompt string, attachment *AiAttachment, histories []AiMessage) (*AiMessage, error) {
	if g.model == "" {
		return nil, fmt.Errorf("model is required")
	}
	if g.client == nil {
		return nil, fmt.Errorf("client is required")
	}

	fmt.Printf("SEND PROMPT %s with OPENAI(%s)\n", prompt, g.model)
	client := *g.client

	var messages []openai.ChatCompletionMessageParamUnion

	if g.systemInstruction != "" {
		messages = append(messages, openai.SystemMessage(g.systemInstruction))
	}

	for _, v := range histories {
		if v.Role == "user" {
			messages = append(messages, openai.UserMessage(v.Content))
		}
		if v.Role == "model" || v.Role == "assistant" {
			messages = append(messages, openai.AssistantMessage(v.Content))
		}
	}

	messages = append(messages, openai.UserMessage(prompt))

	chatReq := openai.ChatCompletionNewParams{
		Messages: messages,
		Model:    g.model,
		ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONObject: &shared.ResponseFormatJSONObjectParam{
				Type: "json_object",
			},
		},
	}
	resp, err := client.Chat.Completions.New(*g.ctx, chatReq)
	if err != nil {
		fmt.Println("Error =>", err)
		return nil, err
	}

	var responseData AiMessage = AiMessage{
		Role:    "model",
		Content: resp.Choices[0].Message.Content,
	}

	fmt.Println("USAGE TOKEN", resp.Usage.TotalTokens)

	return &responseData, nil
}

func (g *OpenAiService) SetApiKey(apiKey string) {
	if apiKey != "" {
		g.ApiKey = apiKey
	}
}

func (g *OpenAiService) SetSystemInstruction(instruction string) {
	if instruction != "" {
		g.systemInstruction = instruction
	}

}

func (g *OpenAiService) SetModel(model string) {
	g.model = model
}
func (g *OpenAiService) SetContentConfig(config *ContentConfig) {

}

func (g *OpenAiService) SetHost(host string) {

}
