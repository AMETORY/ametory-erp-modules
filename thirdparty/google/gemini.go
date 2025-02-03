package google

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

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

func (s *GeminiService) RefreshHistories() {
	getHistories(*s.ctx.Ctx, s)
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
	if strings.HasPrefix(path, "http") {
		resp, err := http.Get(path)
		if err != nil {
			log.Fatalf("Error downloading file: %v", err)
		}
		defer resp.Body.Close()

		file, err := os.CreateTemp("tmp", "gemini-")
		if err != nil {
			log.Fatalf("Error creating temporary file: %v", err)
		}
		defer os.Remove(file.Name())

		_, err = io.Copy(file, resp.Body)
		if err != nil {
			log.Fatalf("Error copying file: %v", err)
		}

		path = file.Name()
	}
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
func (service *GeminiService) GenerateContent(ctx context.Context, input string, userHistories []map[string]interface{}, fileURL, mimeType string) (string, error) {
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
	// session := model.StartChat()

	histories := service.histories
	for _, v := range userHistories {
		role, ok := v["role"].(string)
		if !ok {
			continue
		}
		content, ok := v["content"].(string)
		if !ok {
			continue
		}
		histories = append(histories, &genai.Content{
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

	// for _, v := range session.History {
	// 	fmt.Println(*v)

	// }

	// if session == nil {
	// 	return "", fmt.Errorf("error starting chat session")
	// }
	// fmt.Println("SESSION", session.History)
	// fmt.Println("ctx", ctx)

	// resp, err := session.SendMessage(ctx, genai.Text(input))
	// if err != nil {
	// 	return "", fmt.Errorf("error sending message: %v", err)
	// }
	parts := []genai.Part{}
	for _, part := range histories {
		if part.Role == "user" {
			parts = append(parts, genai.Text("input: "+part.Parts[0].(genai.Text)))
		}
		if part.Role == "model" {
			parts = append(parts, genai.Text("output: "+part.Parts[0].(genai.Text)))
		}

	}

	if fileURL != "" {
		parts = append(parts, genai.FileData{URI: uploadToGemini(ctx, service.client, fileURL, mimeType)})
	}
	parts = append(parts, genai.Text("input: "+input))
	parts = append(parts, genai.Text("output: "))
	// fmt.Println("PARTS", parts)
	// fmt.Println("RESPONSE", resp.Candidates[0].Content)

	resp, err := model.GenerateContent(ctx, parts...)
	if err != nil {
		return "", fmt.Errorf("error generating content: %v", err)
	}
	for _, part := range resp.Candidates[0].Content.Parts {
		return fmt.Sprintf("%v\n", part), nil
	}
	return "", nil
}

func (s *GeminiService) GetHistories() []models.GeminiHistoryModel {
	var historyModels []models.GeminiHistoryModel
	s.ctx.DB.Order("created_at asc").Find(&historyModels)
	return historyModels
}

func (s *GeminiService) UpdateHistory(id string, history models.GeminiHistoryModel) error {

	if err := s.ctx.DB.Where("id = ?", id).Updates(&history).Error; err != nil {
		return fmt.Errorf("error updating history: %v", err)
	}
	return nil
}

func (s *GeminiService) AddHistory(history models.GeminiHistoryModel) error {

	if err := s.ctx.DB.Create(&history).Error; err != nil {
		return fmt.Errorf("error creating history: %v", err)
	}

	return nil
}

func (s *GeminiService) DeleteHistory(id string) error {
	if err := s.ctx.DB.Where("id = ?", id).Delete(&models.GeminiHistoryModel{}).Error; err != nil {
		return fmt.Errorf("error deleting history: %v", err)
	}
	return nil
}

func (s *GeminiService) SetResponseMIMEType(mimetype string) {
	s.responseMimetype = mimetype
}
