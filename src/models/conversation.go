package models

type ConversationMessage struct {
	ID        string                 `json:"id"`
	Role      string                 `json:"role"`
	Content   string                 `json:"content"`
	Intent    string                 `json:"intent,omitempty"`
	Entities  map[string]interface{} `json:"entities,omitempty"`
	CreatedAt int64                  `json:"created_at"`
}

type ConversationHistory struct {
	ID           string                `json:"id"`
	UserID       string                `json:"user_id"`
	Messages     []ConversationMessage `json:"messages"`
	CreatedAt    int64                 `json:"created_at"`
	LastAccessed int64                 `json:"last_accessed"`
}

type IntentPattern struct {
	Name     string   `json:"name"`
	Patterns []string `json:"patterns"`
	Keywords []string `json:"keywords"`
	Action   string   `json:"action"`
}
