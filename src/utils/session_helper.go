// internal/utils/session_helpers.go
package utils

import (
	"net/http"

	"github.com/banking/ai-agents-banking/src/middleware"
	"github.com/banking/ai-agents-banking/src/models"
)

// GetUserSessionFromContext retrieves the user session from the request context
func GetUserSessionFromContext(r *http.Request) (*models.UserSession, bool) {
	session, ok := middleware.GetSessionFromContext(r)
	if !ok {
		return nil, false
	}
	userSession, ok := session.(*models.UserSession)
	return userSession, ok
}

// GetUserIDFromContext retrieves the user ID from the request context
func GetUserIDFromContext(r *http.Request) (userID string, ok bool) {
	// Try to get from middleware context first
	userID, ok = middleware.GetUserIDFromContext(r)
	return
}

// GetUserIDFromRequest extracts user ID from request context with fallback
func GetUserIDFromRequest(r *http.Request) (string, bool) {
	// Try to get from middleware context first
	if userID, ok := middleware.GetUserIDFromContext(r); ok {
		return userID, true
	}

	// Fallback: extract from session
	if session, ok := GetUserSessionFromContext(r); ok {
		return session.GetAccountID(), true
	}

	return "", false
}

// SessionInfo represents session information for API responses
type SessionInfo struct {
	SessionID string `json:"session_id"`
	UserID    string `json:"user_id"`
	ExpiresAt int64  `json:"expires_at"`
}

// GetSessionInfo extracts session information for API responses
func GetSessionInfo(r *http.Request) (*SessionInfo, bool) {
	session, ok := GetUserSessionFromContext(r)
	if !ok {
		return nil, false
	}

	return &SessionInfo{
		SessionID: session.GetID(),
		UserID:    session.GetAccountID(),
		ExpiresAt: session.ExpiresAt.Unix(),
	}, true
}
