package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// LoggingMiddleware logs the request method, path, status, duration, etc.
type LoggingMiddleware struct{}

func NewLoggingMiddleware() *LoggingMiddleware {
	return &LoggingMiddleware{}
}

func (m *LoggingMiddleware) MiddlewareFunc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Use wrapped ResponseWriter that preserves Flusher
		wrapper := &loggingResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Get route info (if using Gorilla Mux)
		route := mux.CurrentRoute(r)
		var routeName string
		if route != nil {
			if name := route.GetName(); name != "" {
				routeName = name
			} else if pathTemplate, err := route.GetPathTemplate(); err == nil {
				routeName = pathTemplate
			}
		}

		next.ServeHTTP(wrapper, r)

		duration := time.Since(start)
		log.Printf(
			"%s %s [%s] %d %v %s %s - %s",
			r.Method,
			r.URL.Path,
			routeName,
			wrapper.statusCode,
			duration,
			r.RemoteAddr,
			r.UserAgent(),
			getStatusText(wrapper.statusCode),
		)
	})
}

// --- Wrapped ResponseWriter with Flusher support ---

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (rw *loggingResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *loggingResponseWriter) Write(data []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(data)
	rw.size += size
	return size, err
}

// Preserve Flusher for streaming support
func (rw *loggingResponseWriter) Flush() {
	if f, ok := rw.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Optional: implement Hijacker, Pusher, ReaderFrom if needed in future

func getStatusText(code int) string {
	switch {
	case code >= 200 && code < 300:
		return "SUCCESS"
	case code >= 300 && code < 400:
		return "REDIRECT"
	case code >= 400 && code < 500:
		return "CLIENT_ERROR"
	case code >= 500:
		return "SERVER_ERROR"
	default:
		return "UNKNOWN"
	}
}
