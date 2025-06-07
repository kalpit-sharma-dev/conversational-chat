package models

// All response types are now defined in agent.go.

type AuthResponse struct {
	Token     string `json:"token"`
	SessionID string `json:"session_id"`
	ExpiresAt int64  `json:"expires_at"`
}

type LlamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}
