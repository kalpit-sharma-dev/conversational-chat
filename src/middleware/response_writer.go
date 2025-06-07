package middleware

import (
	"net/http"
)

// ResponseWriter wraps http.ResponseWriter and captures the status code.
// It also preserves http.Flusher and other interfaces like http.Hijacker (if needed).
type ResponseWriter struct {
	http.ResponseWriter // Embed directly to preserve interfaces
	statusCode          int
}

// NewResponseWriter creates a new wrapped ResponseWriter
func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

// WriteHeader captures the status code
func (rw *ResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// StatusCode returns the captured status code
func (rw *ResponseWriter) StatusCode() int {
	return rw.statusCode
}

// Flush delegates to the underlying ResponseWriter if it supports Flusher
func (rw *ResponseWriter) Flush() {
	if flusher, ok := rw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}
