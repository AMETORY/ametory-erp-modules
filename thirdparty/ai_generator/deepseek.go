package ai_generator

import (
	"context"
	"fmt"

	"github.com/go-deepseek/deepseek"
	"github.com/go-deepseek/deepseek/request"
)

type DeepSeekService struct {
	ctx               *context.Context
	ApiKey            string
	client            *deepseek.Client
	systemInstruction string
	model             string
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
	}
}

func (g *DeepSeekService) Generate(prompt string, attachment *AiAttachment, histories []AiMessage) (*AiMessage, error) {
	if g.model == "" {
		return nil, fmt.Errorf("model is required")
	}
	if g.client == nil {
		return nil, fmt.Errorf("client is required")
	}

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
	resp, err := client.CallChatCompletionsChat(*g.ctx, chatReq)
	if err != nil {
		fmt.Println("Error =>", err)
		return nil, err
	}

	var responseData AiMessage = AiMessage{
		Role:    "model",
		Content: resp.Choices[0].Message.Content,
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

}

func (g *DeepSeekService) SetHost(host string) {

}
