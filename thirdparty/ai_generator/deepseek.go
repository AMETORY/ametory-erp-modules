package ai_generator

import (
	"context"
	"fmt"
	"strings"

	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/go-deepseek/deepseek"
	"github.com/go-deepseek/deepseek/request"
)

type DeepSeekService struct {
	ctx               *context.Context
	ApiKey            string
	client            *deepseek.Client
	systemInstruction string
	model             string
	isJson            bool
}

func NewDeepSeekService(ctx *context.Context, apiKey string) *DeepSeekService {
	client, err := deepseek.NewClient(apiKey)
	if err != nil {
		panic(err)
	}
	return &DeepSeekService{
		ctx:    ctx,
		ApiKey: apiKey,
		client: &client,
		model:  deepseek.DEEPSEEK_CHAT_MODEL,
		isJson: true,
	}
}

func (g *DeepSeekService) Generate(prompt string, attachment *AiAttachment, histories []AiMessage) (*AiMessage, error) {
	if g.model == "" {
		return nil, fmt.Errorf("model is required")
	}
	if g.client == nil {
		return nil, fmt.Errorf("client is required")
	}
	fmt.Printf("SEND PROMPT %s with DEEPSEEK\n", prompt)
	client := *g.client

	var messages []*request.Message

	if g.systemInstruction != "" {
		messages = append(messages, &request.Message{
			Role:    "system",
			Content: g.systemInstruction,
		})
	}
	for _, v := range histories {
		messages = append(messages, &request.Message{
			Role:    v.Role,
			Content: v.Content,
		})
	}

	messages = append(messages, &request.Message{
		Role:    "user",
		Content: prompt,
	})

	chatReq := &request.ChatCompletionsRequest{
		Model:    g.model,
		Messages: messages,
		Stream:   false,
	}

	if g.isJson {
		chatReq.ResponseFormat = &request.ResponseFormat{
			Type: "json_object",
		}
	}

	resp, err := client.CallChatCompletionsChat(*g.ctx, chatReq)
	if err != nil {
		fmt.Println("Error =>", err)
		return nil, err
	}

	fmt.Println("USAGE TOKEN", resp.Usage.TotalTokens)
	utils.LogJson(resp.Usage)

	var responseData AiMessage = AiMessage{
		Role:            "model",
		Content:         resp.Choices[0].Message.Content,
		TotalTokenCount: int32(resp.Usage.TotalTokens),
	}

	return &responseData, nil
}

func (g *DeepSeekService) SetApiKey(apiKey string) {
	if apiKey != "" {
		g.ApiKey = apiKey
	}
}

func (g *DeepSeekService) SetSystemInstruction(instruction string) {
	if instruction != "" {
		g.systemInstruction = instruction
	}

}

func (g *DeepSeekService) SetModel(model string) {
	g.model = model
}
func (g *DeepSeekService) SetContentConfig(config *ContentConfig) {
	if config != nil {
		if strings.Contains(config.ResponseMIMEType, "json") {
			g.isJson = true
		}
		if strings.Contains(config.ResponseMIMEType, "text") {
			g.isJson = false
		}
	}
}

func (g *DeepSeekService) SetHost(host string) {

}
