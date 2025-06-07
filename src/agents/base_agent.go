package agents

import "github.com/banking/ai-agents-banking/src/models"

type BankingAgent interface {
	GetName() string
	GetDescription() string
	CanHandle(intent string, message string) bool
	Process(ctx *models.AgentContext) *models.AgentResponse
	ValidateParameters(params map[string]interface{}) []string
	GetRequiredParameters() []string
	GetHelp() string
	GetTools() []string
	GetConfidence() float64
}

// BaseAgent provides common functionality for all agents
type BaseAgent struct {
	Name        string
	Description string
	Tools       []string
	Confidence  float64
}

func (a *BaseAgent) GetName() string {
	return a.Name
}

func (a *BaseAgent) GetDescription() string {
	return a.Description
}

func (a *BaseAgent) GetTools() []string {
	return a.Tools
}

func (a *BaseAgent) GetConfidence() float64 {
	return a.Confidence
}

func (a *BaseAgent) ValidateParameters(params map[string]interface{}) []string {
	return []string{}
}

func (a *BaseAgent) GetRequiredParameters() []string {
	return []string{}
}

func (a *BaseAgent) GetHelp() string {
	return "No help available for this agent."
}
