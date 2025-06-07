package models

type AgentContext struct {
	SessionID    string
	UserID       string
	Message      string
	Intent       string
	Entities     map[string]interface{}
	Parameters   map[string]interface{}
	Conversation *Conversation
	CurrentStep  *ConversationStep
	Confidence   float64
}

type AgentResponse struct {
	Message           string
	AgentName         string
	Data              interface{}
	Actions           []string
	RequiresInput     bool
	RequiresTool      bool
	ToolName          string
	ToolParams        map[string]interface{}
	MissingParameters []string
}

type Conversation struct {
	ID       string
	Messages []Message
	Context  map[string]interface{}
}

type Message struct {
	Role      string
	Content   string
	Intent    string
	Actions   []string
	Entities  map[string]interface{}
	AgentName string
}

type ConversationStep struct {
	StepID     string
	Intent     string
	Parameters map[string]interface{}
	Missing    []string
	Complete   bool
	AgentName  string
}

type StreamingContext struct {
	Message      string
	Intent       *Intent
	Conversation *Conversation
	Session      *UserSession
}

type Intent struct {
	Name       string
	Confidence float64
	Entities   map[string]interface{}
}

type ChatResponse struct {
	Response    string
	SessionID   string
	Intent      string
	Confidence  float64
	Tools       []string
	AgentUsed   string
	AgentData   interface{}
	NextActions []string
	Entities    map[string]interface{}
}
