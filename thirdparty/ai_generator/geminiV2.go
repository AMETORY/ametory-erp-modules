package ai_generator

import (
	"context"
	"fmt"
	"strings"

	"github.com/AMETORY/ametory-erp-modules/utils"
	"google.golang.org/genai"
)

type GeminiV2SService struct {
	ctx               *context.Context
	ApiKey            string
	client            *genai.Client
	model             string
	systemInstruction string
	contentConfig     *genai.GenerateContentConfig
	isJson            bool
}

func NewGeminiV2Service(ctx *context.Context, apiKey string) *GeminiV2SService {

	client, err := genai.NewClient(*ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})

	if err != nil {
		panic(err)
	}

	return &GeminiV2SService{
		ctx:           ctx,
		ApiKey:        apiKey,
		client:        client,
		model:         "gemini-1.5-flash",
		contentConfig: &genai.GenerateContentConfig{},
		isJson:        true,
	}
}

func (g *GeminiV2SService) Generate(prompt string, attachment *AiAttachment, histories []AiMessage) (*AiMessage, error) {
	if g.model == "" {
		return nil, fmt.Errorf("model is required")
	}
	if g.client == nil {
		return nil, fmt.Errorf("client is required")
	}

	fmt.Printf("SEND PROMPT %s with GEMINI(%s)\n", prompt, g.model)
	var contents []*genai.Content
	for _, v := range histories {
		content := genai.Content{
			Role: v.Role,
			Parts: []*genai.Part{
				{
					Text: v.Content,
				},
			},
		}
		if v.Attachment != nil {
			content.Parts = append(content.Parts, &genai.Part{
				InlineData: &genai.Blob{
					MIMEType: v.Attachment.MimeType,
					Data:     v.Attachment.Data,
				},
			})
		}
		contents = append(contents, &content)
	}

	promptContent := genai.Content{
		Role: "user",
		Parts: []*genai.Part{
			{
				Text: prompt,
			},
		},
	}

	contents = append(contents, &promptContent)
	if attachment != nil {
		promptContent.Parts = append(promptContent.Parts, &genai.Part{
			InlineData: &genai.Blob{
				MIMEType: attachment.MimeType,
				Data:     attachment.Data,
			},
		})
	}

	config := g.contentConfig
	if g.systemInstruction != "" {
		config.SystemInstruction = &genai.Content{
			Role: "system",
			Parts: []*genai.Part{
				{
					Text: g.systemInstruction,
				},
			},
		}
	}
	if g.isJson {
		config.ResponseMIMEType = "application/json"
	}
	resp, err := g.client.Models.GenerateContent(*g.ctx, g.model, contents, config)
	if err != nil {
		return nil, err
	}

	fmt.Println("USAGE TOKEN", resp.UsageMetadata.TotalTokenCount)
	utils.LogJson(resp.UsageMetadata)

	var responseData AiMessage = AiMessage{
		Role:    "model",
		Content: resp.Text(),
	}

	return &responseData, nil
}

func (g *GeminiV2SService) SetApiKey(apiKey string) {
	g.ApiKey = apiKey

}

func (g *GeminiV2SService) SetSystemInstruction(instruction string) {
	if instruction != "" {
		g.systemInstruction = instruction
	}
}

func (g *GeminiV2SService) SetHost(host string) {

}
func (g *GeminiV2SService) SetModel(model string) {
	g.model = model
}
func (g *GeminiV2SService) SetContentConfig(config *ContentConfig) {
	if config != nil {
		g.contentConfig = &genai.GenerateContentConfig{
			Temperature:      config.Temperature,
			TopP:             config.TopP,
			TopK:             config.TopK,
			CandidateCount:   config.CandidateCount,
			MaxOutputTokens:  config.MaxOutputTokens,
			StopSequences:    config.StopSequences,
			ResponseLogprobs: config.ResponseLogprobs,
			Logprobs:         config.Logprobs,
		}

		if strings.Contains(config.ResponseMIMEType, "json") {
			g.isJson = true
		}
		if strings.Contains(config.ResponseMIMEType, "text") {
			g.isJson = false
		}
	}
}
