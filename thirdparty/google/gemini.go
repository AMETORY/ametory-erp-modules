package google

import (
	"context"
	"fmt"
	"log"
	"os"

	erpContext "github.com/AMETORY/ametory-erp-modules/context"
	"github.com/AMETORY/ametory-erp-modules/shared/models"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
	"gorm.io/gorm"
)

type GeminiService struct {
	apiKey             string
	ctx                *erpContext.ERPContext
	histories          []*genai.Content
	client             *genai.Client
	setTemperature     float32
	setTopK            int32
	setTopP            float32
	setMaxOutputTokens int32
	responseMimetype   string
	model              string
	systemInstruction  string
}

func NewGeminiService(ctx *erpContext.ERPContext, apiKey string) *GeminiService {
	if !ctx.SkipMigration {
		Migrate(ctx.DB)
	}
	client, err := genai.NewClient(*ctx.Ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Panicf("Error creating client: %v", err)
	}
	// defer client.Close()

	service := GeminiService{
		apiKey:             apiKey,
		ctx:                ctx,
		setTemperature:     1,
		setTopK:            40,
		setTopP:            0.95,
		setMaxOutputTokens: 8192,
		responseMimetype:   "text/plain",
		model:              "gemini-1.5-flash",
		client:             client,
	}

	getHistories(*ctx.Ctx, &service)
	return &service
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&models.GeminiHistoryModel{})
}

func getHistories(ctx context.Context, service *GeminiService) {
	var historyModels []models.GeminiHistoryModel
	service.ctx.DB.Find(&historyModels)
	// parts := []genai.Part{}
	histories := []*genai.Content{}
	for _, v := range historyModels {
		if v.FileURL != "" {
			histories = append(histories, &genai.Content{
				Role: "user",
				Parts: []genai.Part{
					genai.FileData{URI: uploadToGemini(ctx, service.client, v.FileURL, v.MimeType)},
					genai.Text(v.Input + "\n"),
				},
			})
		} else {
			histories = append(histories, &genai.Content{
				Role: "user",
				Parts: []genai.Part{
					genai.Text(v.Input + "\n"),
				},
			})
		}
		histories = append(histories, &genai.Content{
			Role: "model",
			Parts: []genai.Part{
				genai.Text(v.Output + "\n"),
			},
		})

	}
	service.histories = histories
}

func (service *GeminiService) SetUpSystemInstruction(systemInstruction string) {
	service.systemInstruction = systemInstruction
}

func uploadToGemini(ctx context.Context, client *genai.Client, path, mimeType string) string {
	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}
	defer file.Close()

	options := genai.UploadFileOptions{
		DisplayName: path,
		MIMEType:    mimeType,
	}
	fileData, err := client.UploadFile(ctx, "", file, &options)
	if err != nil {
		log.Fatalf("Error uploading file: %v", err)
	}

	log.Printf("Uploaded file %s as: %s", fileData.DisplayName, fileData.URI)
	return fileData.URI
}

func (service *GeminiService) SetupModel(
	setTemperature float32,
	setTopK int32,
	setTopP float32,
	setMaxOutputTokens int32,
	responseMimetype string,
	model string,
) {

	service.setTemperature = setTemperature
	service.setTopK = setTopK
	service.setTopP = setTopP
	service.setMaxOutputTokens = setMaxOutputTokens
	service.responseMimetype = responseMimetype
	service.model = model

}
func (service *GeminiService) GenerateContent(ctx context.Context, input string, userHistories []map[string]interface{}) (string, error) {
	if service.client == nil {
		return "", fmt.Errorf("client is not initialized")
	}
	model := service.client.GenerativeModel(service.model)
	if model == nil {
		return "", fmt.Errorf("model is not found")
	}

	model.SetTemperature(service.setTemperature)
	model.SetTopK(service.setTopK)
	model.SetTopP(service.setTopP)
	model.SetMaxOutputTokens(service.setMaxOutputTokens)
	model.ResponseMIMEType = service.responseMimetype
	session := model.StartChat()

	session.History = append(session.History, service.histories...)
	for _, v := range userHistories {
		role, ok := v["role"].(string)
		if !ok {
			continue
		}
		content, ok := v["content"].(string)
		if !ok {
			continue
		}
		session.History = append(session.History, &genai.Content{
			Role: role,
			Parts: []genai.Part{
				genai.Text(content + "\n"),
			},
		})
	}
	if service.systemInstruction != "" {
		model.SystemInstruction = &genai.Content{
			Parts: []genai.Part{genai.Text(service.systemInstruction)},
		}
	}

	if session == nil {
		return "", fmt.Errorf("error starting chat session")
	}
	// fmt.Println("SESSION", session)
	// fmt.Println("ctx", ctx)

	resp, err := session.SendMessage(ctx, genai.Text(input))
	if err != nil {
		return "", fmt.Errorf("error sending message: %v", err)
	}

	for _, part := range resp.Candidates[0].Content.Parts {
		return fmt.Sprintf("%v\n", part), nil
	}
	return "", nil
}
