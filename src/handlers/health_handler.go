package handlers

import (
	"encoding/json"
	"net/http"
	"runtime"
	"time"

	"github.com/banking/ai-agents-banking/src/services"
)

type HealthHandler struct {
	agentService        *services.AgentService
	conversationService *services.ConversationService
	sessionService      *services.SessionService
}

func NewHealthHandler(
	agentService *services.AgentService,
	conversationService *services.ConversationService,
	sessionService *services.SessionService,
) *HealthHandler {
	return &HealthHandler{
		agentService:        agentService,
		conversationService: conversationService,
		sessionService:      sessionService,
	}
}

func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Get system metrics
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	status := map[string]interface{}{
		"status":    "OK",
		"timestamp": time.Now(),
		"version":   "1.0.0",
		"service":   "Banking Agents API",
		"metrics": map[string]interface{}{
			"agents":          h.agentService.GetAgentsCount(),
			"active_sessions": h.sessionService.GetActiveSessionsCount(),
			"conversations":   h.conversationService.GetActiveConversationsCount(),
			"memory": map[string]interface{}{
				"alloc":       m.Alloc,
				"total_alloc": m.TotalAlloc,
				"sys":         m.Sys,
				"num_gc":      m.NumGC,
			},
			"goroutines": runtime.NumGoroutine(),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}
