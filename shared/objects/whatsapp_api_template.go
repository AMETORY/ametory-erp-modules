package objects

type WhatsappAPITemplateResponse struct {
	Data   []WhatsappAPITemplate `json:"data"`
	Paging Paging                `json:"paging"`
}

type WhatsappAPITemplate struct {
	Name            string      `json:"name"`
	ParameterFormat string      `json:"parameter_format"`
	Components      []Component `json:"components"`
	Language        string      `json:"language"`
	Status          string      `json:"status"`
	Category        string      `json:"category"`
	SubCategory     string      `json:"sub_category,omitempty"`
	ID              string      `json:"id"`
}

type Component struct {
	Type    string    `json:"type"`
	Format  string    `json:"format,omitempty"`
	Text    string    `json:"text,omitempty"`
	Example *Example  `json:"example,omitempty"`
	Buttons []Buttons `json:"buttons,omitempty"`
}

type Example struct {
	HeaderHandle        []string              `json:"header_handle,omitempty"`
	BodyText            [][]string            `json:"body_text,omitempty"`
	BodyTextNamedParams []BodyTextNamedParams `json:"body_text_named_params,omitempty"`
}

type BodyTextNamedParams struct {
	ParamName string `json:"param_name"`
	Example   string `json:"example"`
}

type Buttons struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type Paging struct {
	Cursors Cursors `json:"cursors"`
}

type Cursors struct {
	Before string `json:"before"`
	After  string `json:"after"`
}
