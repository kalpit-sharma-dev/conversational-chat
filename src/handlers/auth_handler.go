package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/banking/ai-agents-banking/src/models"
	"github.com/banking/ai-agents-banking/src/services"
)

type AuthHandler struct {
	sessionService *services.SessionService
}

func NewAuthHandler(sessionService *services.SessionService) *AuthHandler {
	return &AuthHandler{
		sessionService: sessionService,
	}
}

func (h *AuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		UserID   string `json:"user_id"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// TODO: Add proper user authentication here
	// For now, we'll just validate that user_id is provided
	if req.UserID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// Create new session for user
	session, err := h.sessionService.CreateUserSession(req.UserID)
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	response := models.AuthResponse{
		Token:     session.Token,
		SessionID: session.ID,
		ExpiresAt: session.ExpiresAt.Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
