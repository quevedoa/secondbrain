package entity

type LLMRequest struct {
	Model   string   `json:"model"`
	Prompts []Prompt `json:"prompts"`
	Stream  bool     `json:"stream"`
}

type LLMResponse struct {
	Responses []string `json:"responses"`
}

type Prompt struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type OllamaResponse struct {
	Model      string `json:"model"`
	Response   string `json:"response"`
	Done       bool   `json:"done"`
	DoneReason string `json:"done_reason"`
}
