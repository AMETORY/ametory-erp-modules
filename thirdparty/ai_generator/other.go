package ai_generator

import (
	"context"
)

type OtherService struct {
	ctx    *context.Context
	ApiKey string
}

func NewOtherService(ctx *context.Context, apiKey string) *OtherService {
	return &OtherService{
		ctx:    ctx,
		ApiKey: apiKey,
	}
}

func (g *OtherService) Generate(prompt string, attachment *AiAttachment, histories []AiMessage) (*AiMessage, error) {

	var responseData AiMessage = AiMessage{
		Role: "model",
	}

	return &responseData, nil
}

func (g *OtherService) SetApiKey(apiKey string) {

}

func (g *OtherService) SetSystemInstruction(instruction string) {

}

func (g *OtherService) SetModel(model string) {
}
func (g *OtherService) SetContentConfig(config *ContentConfig) {

}

func (g *OtherService) SetHost(host string) {

}
