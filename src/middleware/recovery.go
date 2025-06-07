package middleware

import (
	"log"
	"net/http"
	"runtime/debug"
)

type RecoveryMiddleware struct{}

func NewRecoveryMiddleware() *RecoveryMiddleware {
	return &RecoveryMiddleware{}
}

func (m *RecoveryMiddleware) MiddlewareFunc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v\n%s", err, debug.Stack())

				w.Header().Set("Content-Type", "application/json")
				http.Error(w, `{"error": "Internal server error", "code": "INTERNAL_ERROR"}`, http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
