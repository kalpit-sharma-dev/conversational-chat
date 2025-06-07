package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/banking/ai-agents-banking/src/services"
	"github.com/banking/ai-agents-banking/src/utils"
	"github.com/gorilla/mux"
)

type AgentsHandler struct {
	agentService   *services.AgentService
	sessionService *services.SessionService
}

func NewAgentsHandler(agentService *services.AgentService, sessionService *services.SessionService) *AgentsHandler {
	return &AgentsHandler{
		agentService:   agentService,
		sessionService: sessionService,
	}
}

func (h *AgentsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get session info using helper
	session, ok := utils.GetUserSessionFromContext(r)
	if !ok {
		http.Error(w, "Session not found in context", http.StatusUnauthorized)
		return
	}

	// Get all registered agents
	agents := h.agentService.GetAllAgents()

	agentInfo := make(map[string]interface{})
	for name, agent := range agents {
		agentInfo[name] = map[string]interface{}{
			"name":        agent.GetName(),
			"description": agent.GetDescription(),
			"help":        agent.GetHelp(),
			"capabilities": map[string]interface{}{
				"can_transfer":      agent.CanHandle("transfer", ""),
				"can_check_balance": agent.CanHandle("balance", ""),
				"can_add_payee":     agent.CanHandle("add_payee", ""),
				"can_apply_loan":    agent.CanHandle("loan", ""),
			},
		}
	}

	response := map[string]interface{}{
		"agents":     agentInfo,
		"total":      len(agentInfo),
		"session_id": session.ID,
		"user_id":    session.AccountID,
		"timestamp":  time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *AgentsHandler) GetAgentDetails(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	agentName := vars["agentName"]

	// Get session info using helper
	session, ok := utils.GetUserSessionFromContext(r)
	if !ok {
		http.Error(w, "Session not found in context", http.StatusUnauthorized)
		return
	}

	// Get all registered agents
	agents := h.agentService.GetAllAgents()

	// Find specific agent
	var agentInfo map[string]interface{}
	for name, agent := range agents {
		if name == agentName || agent.GetName() == agentName {
			agentInfo = map[string]interface{}{
				"name":            agent.GetName(),
				"description":     agent.GetDescription(),
				"help":            agent.GetHelp(),
				"required_params": agent.GetRequiredParameters(),
				"capabilities": map[string]interface{}{
					"can_transfer":      agent.CanHandle("transfer", ""),
					"can_check_balance": agent.CanHandle("balance", ""),
					"can_add_payee":     agent.CanHandle("add_payee", ""),
					"can_apply_loan":    agent.CanHandle("loan", ""),
				},
				"session_id": session.ID,
				"user_id":    session.AccountID,
				"timestamp":  time.Now(),
			}
			break
		}
	}

	if agentInfo == nil {
		http.Error(w, `{"error": "Agent not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(agentInfo)
}
