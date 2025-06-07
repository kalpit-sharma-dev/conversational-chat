package services

import "time"

// Session represents a chat session
type Session struct {
	ID        string
	CreatedAt time.Time
	// Add other session-related fields as needed
}

// SessionService handles session management
type SessionService interface {
	GetSession(sessionID string) (*Session, error)
	CreateSession() (*Session, error)
	DeleteSession(sessionID string) error
}

// AgentService handles agent management
type AgentService interface {
	GetAgentForIntent(intent string) (*Agent, error)
	ListAgents() ([]*Agent, error)
}

// Agent represents an AI agent
type Agent struct {
	ID           string
	Name         string
	Description  string
	Capabilities []string
}

// ConversationService handles conversation history
type ConversationService interface {
	AddMessage(sessionID, userMessage, assistantMessage string) error
	GetMessages(sessionID string) ([]*Message, error)
}

// Message represents a chat message
type Message struct {
	ID        string
	SessionID string
	Role      string // "user" or "assistant"
	Content   string
	Timestamp time.Time
}

// IntentRecognitionService handles intent recognition
type IntentRecognitionService interface {
	RecognizeIntent(message string) (string, error)
}

// LlamaService handles LLM interactions
type LlamaService interface {
	GenerateResponse(message string, agent *Agent) (string, error)
}

// ToolRegistry manages available tools
type ToolRegistry interface {
	GetTool(name string) (Tool, error)
	ListTools() ([]Tool, error)
}

// Tool represents a capability that can be used by agents
type Tool interface {
	Name() string
	Description() string
	Execute(params map[string]interface{}) (interface{}, error)
}
