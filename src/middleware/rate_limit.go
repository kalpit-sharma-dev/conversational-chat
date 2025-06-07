package middleware

import (
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type RateLimitMiddleware struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}

func NewRateLimitMiddleware(requestsPerSecond int, burst int) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		limiters: make(map[string]*rate.Limiter),
		rate:     rate.Limit(requestsPerSecond),
		burst:    burst,
	}
}

func (m *RateLimitMiddleware) MiddlewareFunc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Use IP address as key
		key := m.getKey(r)

		limiter := m.getLimiter(key)

		if !limiter.Allow() {
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{"error": "Rate limit exceeded", "code": "RATE_LIMIT_EXCEEDED"}`, http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (m *RateLimitMiddleware) getKey(r *http.Request) string {
	// Use IP address as the key
	return r.RemoteAddr
}

func (m *RateLimitMiddleware) getLimiter(key string) *rate.Limiter {
	m.mu.Lock()
	defer m.mu.Unlock()

	limiter, exists := m.limiters[key]
	if !exists {
		limiter = rate.NewLimiter(m.rate, m.burst)
		m.limiters[key] = limiter

		// Clean up old limiters periodically
		go func() {
			time.Sleep(time.Hour)
			m.mu.Lock()
			delete(m.limiters, key)
			m.mu.Unlock()
		}()
	}

	return limiter
}
