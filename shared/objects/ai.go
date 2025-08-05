package objects

type AiResponse struct {
	Response string      `json:"response"`
	Type     string      `json:"type"`
	Command  string      `json:"command"`
	Params   interface{} `json:"params"`
}
