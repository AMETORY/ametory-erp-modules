package models

type ChatbotFlow struct {
	Flows                map[string]Flow   `json:"flows" bson:"flows"`
	InitialFlow          string            `json:"initial_flow" bson:"initialFlow"`
	Keywords             map[string]string `json:"keywords" bson:"keywords"`
	FallbackResponse     string            `json:"fallback_response" bson:"fallbackResponse"`
	FallbackResponseType string            `json:"fallback_response_type" bson:"fallbackResponseType"`
	RefID                string            `json:"ref_id" bson:"refId"`
}

type Flow struct {
	Type              string        `json:"type" bson:"type"`
	Text              string        `json:"text" bson:"text"`
	Options           []Option      `json:"options" bson:"options"`
	DefaultResponse   string        `json:"default_response" bson:"defaultResponse"`
	Steps             []Step        `json:"steps" bson:"steps"`
	CompletionMessage string        `json:"completion_message" bson:"completionMessage"`
	BackToFlow        string        `json:"back_to_flow" bson:"backToFlow"`
	Agent             *AiAgentModel `json:"agent" bson:"agent,omitempty" gorm:"-"`
	AgentID           *string       `json:"agent_id" bson:"agentId"`
}

type Option struct {
	Input    string `json:"input" bson:"input"`
	Display  string `json:"display" bson:"display"`
	NextFlow string `json:"next_flow" bson:"nextFlow"`
	Response string `json:"response" bson:"response"`
}

type Step struct {
	Field      string     `json:"field" bson:"field"`
	Question   string     `json:"question" bson:"question"`
	Validation Validation `json:"validation" bson:"validation"`
}

type Validation struct {
	Type         string `json:"type" bson:"type"`
	Value        string `json:"value" bson:"value"`
	Pattern      string `json:"pattern" bson:"pattern"`
	ErrorMessage string `json:"error_message" bson:"errorMessage"`
}
