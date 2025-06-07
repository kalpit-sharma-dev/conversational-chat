package models

type ChatRequest struct {
	Message   string `json:"message"`
	SessionID string `json:"session_id,omitempty"`
	Token     string `json:"token,omitempty"`
	Stream    bool   `json:"stream,omitempty"`
}

type LlamaRequest struct {
	Model   string                 `json:"model"`
	Prompt  string                 `json:"prompt"`
	Stream  bool                   `json:"stream"`
	Options map[string]interface{} `json:"options"`
	System  string                 `json:"system,omitempty"`
}
