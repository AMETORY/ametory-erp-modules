package ai_generator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/AMETORY/ametory-erp-modules/utils"
)

type OllamaService struct {
	ctx               *context.Context
	host              string
	model             string
	systemInstruction string
	stream            bool
	contentConfig     *ContentConfig
}

func NewOllamaService(ctx *context.Context, model string) *OllamaService {
	return &OllamaService{
		ctx:   ctx,
		host:  "http://localhost:11434",
		model: model,
	}
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Request struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
	Format   string    `json:"format"`
	Options  struct {
		NumCtx int `json:"num_ctx"`
	}
}
type Response struct {
	Model              string    `json:"model"`
	CreatedAt          time.Time `json:"created_at"`
	Message            Message   `json:"message"`
	Done               bool      `json:"done"`
	TotalDuration      int64     `json:"total_duration"`
	LoadDuration       int       `json:"load_duration"`
	PromptEvalCount    int       `json:"prompt_eval_count"`
	PromptEvalDuration int       `json:"prompt_eval_duration"`
	EvalCount          int       `json:"eval_count"`
	EvalDuration       int64     `json:"eval_duration"`
}

func (g *OllamaService) Generate(prompt string, attachment *AiAttachment, histories []AiMessage) (*AiMessage, error) {
	if g.model == "" {
		return nil, fmt.Errorf("ollama model is required")
	}
	messages := []Message{}
	if g.systemInstruction != "" {
		messages = append(messages, Message{
			Role: "system", Content: g.systemInstruction,
		})
	}
	for _, v := range histories {
		messages = append(messages, Message{
			Role:    v.Role,
			Content: v.Content,
		})
	}

	messages = append(messages, Message{
		Role: "user", Content: prompt,
	})
	ollamaReq := Request{
		Model:    g.model,
		Messages: messages,
		Stream:   false,
		Options: struct {
			NumCtx int `json:"num_ctx"`
		}{
			NumCtx: 8192,
		},
	}

	if g.contentConfig != nil {
		cfg := *g.contentConfig
		ollamaReq.Format = cfg.ResponseMIMEType
	}

	utils.LogJson(ollamaReq)

	js, err := json.Marshal(&ollamaReq)
	if err != nil {
		return nil, err
	}
	client := http.Client{}

	url := fmt.Sprintf("%s/api/chat", g.host)
	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(js))
	if err != nil {
		return nil, err
	}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()
	ollamaResp := Response{}
	err = json.NewDecoder(httpResp.Body).Decode(&ollamaResp)
	if err != nil {
		return nil, err
	}

	var responseData AiMessage = AiMessage{
		Role:    "model",
		Content: ollamaResp.Message.Content,
	}

	return &responseData, nil
}

func (g *OllamaService) SetApiKey(apiKey string) {

}
func (g *OllamaService) SetSteam(stream bool) {
	g.stream = stream
}

func (g *OllamaService) SetSystemInstruction(instruction string) {
	if instruction != "" {
		g.systemInstruction = instruction
	}
}

func (g *OllamaService) SetModel(model string) {
	if model == "" {
		g.model = model
	}
}
func (g *OllamaService) SetContentConfig(config *ContentConfig) {
	g.contentConfig = config

}

func (g *OllamaService) SetHost(host string) {
	g.host = host
}
