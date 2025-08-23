package ai_generator

import (
	"context"
	"fmt"
)

type AiGenerator interface {
	Generate(prompt string, attachment *AiAttachment, histories []AiMessage) (*AiMessage, error)
	SetApiKey(apiKey string)
	SetSystemInstruction(instruction string)
	SetModel(model string)
	SetContentConfig(config *ContentConfig)
	SetHost(host string)
}

type AiMessage struct {
	Role            string // Contoh: "user", "model", "system"
	Content         string
	Attachment      *AiAttachment
	TotalTokenCount int32
}

type AiAttachment struct {
	MimeType string
	Data     []byte
}

type GeneratorConfig struct {
	Ctx               *context.Context
	APIKey            string
	SystemInstruction string
	Model             string
	Host              string
}

type ContentConfig struct {
	Temperature      *float32 `json:"temperature,omitempty"`
	TopP             *float32 `json:"topP,omitempty"`
	TopK             *float32 `json:"topK,omitempty"`
	CandidateCount   int32    `json:"candidateCount,omitempty"`
	MaxOutputTokens  int32    `json:"maxOutputTokens,omitempty"`
	StopSequences    []string `json:"stopSequences,omitempty"`
	ResponseLogprobs bool     `json:"responseLogprobs,omitempty"`
	Logprobs         *int32   `json:"logprobs,omitempty"`
	PresencePenalty  *float32 `json:"presencePenalty,omitempty"`
	FrequencyPenalty *float32 `json:"frequencyPenalty,omitempty"`
	Seed             *int32   `json:"seed,omitempty"`
	ResponseMIMEType string   `json:"responseMimeType,omitempty"`
}

type GeneratorFactory func(config GeneratorConfig) (AiGenerator, error)

func NewAiGenerator(factory GeneratorFactory, config GeneratorConfig) (AiGenerator, error) {
	return factory(config)
}

func ExampleNewAiGenerator() {
	factory := func(config GeneratorConfig) (AiGenerator, error) {
		init := NewDeepSeekService(config.Ctx, config.APIKey)
		return init, nil
	}
	config := GeneratorConfig{
		APIKey: "1234567890",
	}
	generator, err := NewAiGenerator(factory, config)
	if err != nil {
		panic(err)
	}

	prompt := "Halo, bagaimana kabar?"
	histories := []AiMessage{
		{
			Role:    "user",
			Content: "Hai, saya ingin tahu tentang cuaca di Bandung.",
		},
	}

	output, err := generator.Generate(prompt, nil, histories)
	if err != nil {
		panic(err)
	}
	fmt.Println(output)
	// Output: Halo, kabar baik. Cuaca di Bandung saat ini cerah.
}
