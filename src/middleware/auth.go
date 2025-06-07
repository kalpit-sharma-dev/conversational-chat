// pkg/middleware/auth.go
package middleware

import (
	"context"
	"net/http"
	"strings"
)

// SessionValidator is an interface to avoid circular import
// We use interface{} to avoid importing models package
type SessionValidator interface {
	ValidateToken(token string) (interface{}, bool)
}

type AuthMiddleware struct {
	sessionValidator SessionValidator
}

func NewAuthMiddleware(sessionValidator SessionValidator) *AuthMiddleware {
	return &AuthMiddleware{
		sessionValidator: sessionValidator,
	}
}

// MiddlewareFunc returns a Gorilla Mux compatible middleware function
func (m *AuthMiddleware) MiddlewareFunc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header or query parameter
		token := m.extractToken(r)

		// Validate token
		session, valid := m.sessionValidator.ValidateToken(token)
		if !valid {
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{"error": "Invalid or expired token", "code": "UNAUTHORIZED"}`, http.StatusUnauthorized)
			return
		}

		// Add session to request context using type assertion
		if sessionObj, ok := session.(interface{ GetAccountID() string }); ok {
			ctx := context.WithValue(r.Context(), "session", session)
			ctx = context.WithValue(ctx, "userID", sessionObj.GetAccountID())
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			// Fallback if type assertion fails
			ctx := context.WithValue(r.Context(), "session", session)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	})
}

// Legacy function for backward compatibility
func (m *AuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := m.extractToken(r)

		_, valid := m.sessionValidator.ValidateToken(token)
		if !valid {
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{"error": "Invalid or expired token", "code": "UNAUTHORIZED"}`, http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (m *AuthMiddleware) extractToken(r *http.Request) string {
	// Try Authorization header first
	token := r.Header.Get("Authorization")
	if strings.HasPrefix(token, "Bearer ") {
		return strings.TrimPrefix(token, "Bearer ")
	}

	// Try query parameter
	if token = r.URL.Query().Get("token"); token != "" {
		return token
	}

	// Try form value (for POST requests)
	if r.Method == http.MethodPost {
		r.ParseForm()
		if token = r.FormValue("token"); token != "" {
			return token
		}
	}

	return ""
}

// GetSessionFromContext extracts session from request context
// Returns interface{} to avoid import issues
func GetSessionFromContext(r *http.Request) (interface{}, bool) {
	if session := r.Context().Value("session"); session != nil {
		return session, true
	}
	return nil, false
}

// GetUserIDFromContext extracts user ID from request context
func GetUserIDFromContext(r *http.Request) (string, bool) {
	if userID, ok := r.Context().Value("userID").(string); ok {
		return userID, true
	}

	// Fallback: try to get from session if userID wasn't set
	if session := r.Context().Value("session"); session != nil {
		if sessionObj, ok := session.(interface{ GetAccountID() string }); ok {
			return sessionObj.GetAccountID(), true
		}
	}

	return "", false
}
