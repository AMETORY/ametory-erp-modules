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
	"github.com/AMETORY/ametory-erp-modules/utils"
	"github.com/google/generative-ai-go/genai"
	"github.com/morkid/paginate"
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
	agentID            *string
	sessionCode        *string
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

func (s *GeminiService) SetupAgentID(agentID string) {
	s.agentID = &agentID
}

func (s *GeminiService) SetupSessionCode(sessionCode string) {
	s.sessionCode = &sessionCode
}

func (s *GeminiService) SetupAPIKey(apiKey string, skipHistory bool) {
	s.apiKey = apiKey

	client, err := genai.NewClient(*s.ctx.Ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Panicf("Error creating client: %v", err)
	}
	// defer client.Close()

	s.client = client

	if !skipHistory {
		s.RefreshHistories()
	} else {
		s.histories = []*genai.Content{}
	}

}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&models.GeminiHistoryModel{}, &models.GeminiAgent{})
}

func (s *GeminiService) RefreshHistories() {
	getHistories(*s.ctx.Ctx, s)
}

func getHistories(ctx context.Context, service *GeminiService) {
	var historyModels []models.GeminiHistoryModel
	db := service.ctx.DB.Model(&models.GeminiHistoryModel{})
	if service.agentID != nil {
		db = db.Where("agent_id = ?", *service.agentID)
	}
	if service.sessionCode != nil {
		db = db.Where("session_code = ?", *service.sessionCode)
	}
	db.Find(&historyModels)
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

func (s *GeminiService) UploadToGemini(path, mimeType string) (string, error) {
	response := uploadToGemini(*s.ctx.Ctx, s.client, path, mimeType)
	if response == "" {
		return "", fmt.Errorf("Error uploading file")
	}
	return response, nil
}
func uploadToGemini(ctx context.Context, client *genai.Client, path, mimeType string) string {
	if strings.HasPrefix(path, "http") {
		resp, err := http.Get(path)
		if err != nil {
			log.Fatalf("Error downloading file: %v", err)
			return ""
		}
		defer resp.Body.Close()

		file, err := os.CreateTemp("tmp", "gemini-")
		if err != nil {
			log.Fatalf("Error creating temporary file: %v", err)
			return ""
		}
		defer os.Remove(file.Name())

		_, err = io.Copy(file, resp.Body)
		if err != nil {
			log.Fatalf("Error copying file: %v", err)
			return ""
		}

		path = file.Name()
	}
	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
		return ""
	}
	defer file.Close()

	options := genai.UploadFileOptions{
		DisplayName: path,
		MIMEType:    mimeType,
	}
	fileData, err := client.UploadFile(ctx, "", file, &options)
	if err != nil {
		log.Fatalf("Error uploading file: %v", err)
		return ""
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

	// fmt.Println(
	// 	setTemperature,
	// 	setTopK,
	// 	setTopP,
	// 	setMaxOutputTokens,
	// 	responseMimetype,
	// 	model,
	// )

	service.setTemperature = setTemperature
	service.setTopK = setTopK
	service.setTopP = setTopP
	service.setMaxOutputTokens = setMaxOutputTokens
	service.responseMimetype = responseMimetype
	service.model = model

}
func (service *GeminiService) GenerateContent(ctx context.Context, input string, userHistories []map[string]any, fileURL, mimeType string) (string, error) {
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

	histories := service.histories
	if fileURL != "" {
		upload := uploadToGemini(ctx, service.client, fileURL, mimeType)
		histories = append(histories, &genai.Content{
			Role: "model",
			Parts: []genai.Part{
				genai.FileData{URI: upload, MIMEType: mimeType},
			},
		})
	}
	for _, v := range userHistories {
		role, ok := v["role"].(string)
		if !ok {
			continue
		}
		content, ok := v["content"].(string)
		if !ok {
			continue
		}

		fileURL, ok := v["file_url"].(string)
		if ok {
			mType, _ := v["mime_type"].(string)
			upload := uploadToGemini(ctx, service.client, fileURL, mType)
			histories = append(histories, &genai.Content{
				Role: role,
				Parts: []genai.Part{
					genai.FileData{URI: upload, MIMEType: mType},
				},
			})
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

	// utils.LogJson(histories)
	// for _, v := range histories {
	// 	fmt.Printf("%s:\t %s\n", v.Role, v.Parts[0].(genai.Text))
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
	// parts := []genai.Part{}
	// for _, part := range histories {
	// 	if part.Role == "user" {
	// 		parts = append(parts, genai.Text("input: "+part.Parts[0].(genai.Text)))
	// 	}
	// 	if part.Role == "model" {
	// 		parts = append(parts, genai.Text("output: "+part.Parts[0].(genai.Text)))
	// 	}

	// }

	// if fileURL != "" {
	// 	parts = append(parts, genai.FileData{URI: uploadToGemini(ctx, service.client, fileURL, mimeType)})
	// }
	// parts = append(parts, genai.Text("input: "+input))
	// parts = append(parts, genai.Text("output: "))
	// fmt.Println("PARTS", parts)
	// fmt.Println("RESPONSE", resp.Candidates[0].Content)

	session.History = histories

	resp, err := session.SendMessage(ctx, genai.Text(input))
	// resp, err := model.GenerateContent(ctx, parts...)
	if err != nil {
		return "", fmt.Errorf("error generating content: %v", err)
	}
	for _, part := range resp.Candidates[0].Content.Parts {
		return fmt.Sprintf("%v\n", part), nil
	}
	return "", nil
}

func (s *GeminiService) GetHistories(agentID *string, companyID *string) []models.GeminiHistoryModel {
	var historyModels []models.GeminiHistoryModel
	db := s.ctx.DB.Model(&models.GeminiHistoryModel{})
	if agentID != nil {
		db = db.Where("agent_id = ?", agentID)
	}
	if companyID != nil {
		db = db.Where("company_id = ?", companyID)
	}
	db.Order("created_at asc").Find(&historyModels)
	return historyModels
}

func (s *GeminiService) UpdateHistory(id string, history models.GeminiHistoryModel) error {

	if err := s.ctx.DB.Where("id = ?", id).Updates(&history).Error; err != nil {
		return fmt.Errorf("error updating history: %v", err)
	}
	return nil
}

func (s *GeminiService) AddHistory(history models.GeminiHistoryModel) error {
	if s.sessionCode != nil {
		history.SessionCode = s.sessionCode
	}
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

func (s *GeminiService) CreateAgent(agent *models.GeminiAgent) error {
	if err := s.ctx.DB.Create(agent).Error; err != nil {
		return fmt.Errorf("error creating agent: %v", err)
	}

	return nil
}

func (s *GeminiService) UpdateAgent(id string, agent *models.GeminiAgent) error {

	if err := s.ctx.DB.Where("id = ?", id).Updates(agent).Error; err != nil {
		return fmt.Errorf("error updating agent: %v", err)
	}
	return nil
}

func (s *GeminiService) DeleteAgent(id string) error {
	if err := s.ctx.DB.Where("id = ?", id).Delete(&models.GeminiAgent{}).Error; err != nil {
		return fmt.Errorf("error deleting agent: %v", err)
	}
	return nil
}

func (s *GeminiService) GetAgents(request http.Request) (paginate.Page, error) {
	pg := paginate.New()
	stmt := s.ctx.DB.Order("created_at desc").Model(&models.GeminiAgent{})
	utils.FixRequest(&request)
	page := pg.With(stmt).Request(request).Response(&[]models.GeminiAgent{})
	page.Page = page.Page + 1
	return page, nil
}

func (s *GeminiService) GetAgent(id string) (*models.GeminiAgent, error) {
	var agent models.GeminiAgent

	if err := s.ctx.DB.Where("id = ?", id).First(&agent).Error; err != nil {
		return nil, fmt.Errorf("error getting agent: %v", err)
	}
	return &agent, nil
}
