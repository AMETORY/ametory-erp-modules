package objects

type AiResponse struct {
	Response        string `json:"response"`
	Type            string `json:"type"`
	Command         string `json:"command"`
	Params          any    `json:"params"`
	TotalTokenCount int32  `json:"total_token_count"`
}
