package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/banking/ai-agents-banking/src/models"
	"github.com/banking/ai-agents-banking/src/services"
	"github.com/banking/ai-agents-banking/src/utils"
	"github.com/gorilla/mux"
)

type ConversationHandler struct {
	conversationService *services.ConversationService
	sessionService      *services.SessionService
}

func NewConversationHandler(conversationService *services.ConversationService, sessionService *services.SessionService) *ConversationHandler {
	return &ConversationHandler{
		conversationService: conversationService,
		sessionService:      sessionService,
	}
}

func (h *ConversationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get session from context
	session, ok := utils.GetUserSessionFromContext(r)
	if !ok {
		http.Error(w, "Session not found in context", http.StatusUnauthorized)
		return
	}

	messages := h.conversationService.GetConversationHistory(session.ID)
	if messages == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"session_id": session.ID,
			"messages":   []models.Message{},
			"context":    map[string]interface{}{},
		})
		return
	}

	// Get conversation memory for additional context
	memory := h.conversationService.GetConversationMemory(session.ID)
	context, ok := memory.GetContext(session.ID)
	if !ok {
		context = make(map[string]interface{})
	}

	response := map[string]interface{}{
		"session_id": session.ID,
		"messages":   messages,
		"context":    context,
		"metadata": map[string]interface{}{
			"total_messages": len(messages),
			"last_updated":   time.Now(),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *ConversationHandler) ClearHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	sessionID := vars["sessionId"]

	// Get session from context
	session, ok := utils.GetUserSessionFromContext(r)
	if !ok {
		http.Error(w, "Session not found in context", http.StatusUnauthorized)
		return
	}

	// Verify session ID matches
	if session.ID != sessionID {
		http.Error(w, "Session ID mismatch", http.StatusForbidden)
		return
	}

	// Clear conversation history
	h.conversationService.ClearConversation(sessionID)

	response := map[string]interface{}{
		"session_id": sessionID,
		"message":    "Conversation history cleared successfully",
		"cleared":    true,
		"timestamp":  time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
