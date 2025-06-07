package models

import (
	"sync"
	"time"
)

type StreamingSession struct {
	ID             string        `json:"id"`
	Token          string        `json:"token"`
	CreatedAt      time.Time     `json:"created_at"`
	LastPoll       time.Time     `json:"last_poll"`
	ExpiresAt      time.Time     `json:"expires_at"`
	Done           bool          `json:"done"`
	Content        string        `json:"content"`
	ContentChannel chan string   `json:"-"` // Not serialized
	mu             sync.RWMutex  `json:"-"` // Not serialized
	subscribers    []chan string `json:"-"` // Not serialized
}

type UserSession struct {
	ID          string
	Token       string
	CreatedAt   time.Time
	ExpiresAt   time.Time
	LastUsed    time.Time
	AccountID   string
	CurrentStep *ConversationStep
	mu          sync.RWMutex
}

// NewStreamingSession creates a new streaming session
func NewStreamingSession(token string) *StreamingSession {
	now := time.Now()
	return &StreamingSession{
		ID:             generateSessionID(),
		Token:          token,
		CreatedAt:      now,
		LastPoll:       now,
		ExpiresAt:      now.Add(30 * time.Minute),
		Done:           false,
		Content:        "",
		ContentChannel: make(chan string, 100), // Buffered channel
		subscribers:    make([]chan string, 0),
	}
}

// AppendContent adds new content to the session
func (s *StreamingSession) AppendContent(content string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Done {
		return // Don't append if already done
	}

	// Only send new content
	s.Content += content
	s.LastPoll = time.Now()

	// Send to buffered channel (non-blocking)
	select {
	case s.ContentChannel <- content: // Send only the new content
	default:
		// Channel is full, skip this content chunk
	}

	// Notify all subscribers with only new content
	for _, subscriber := range s.subscribers {
		select {
		case subscriber <- content: // Send only the new content
		default:
			// Subscriber channel is full, skip
		}
	}
}

// GetContentAndDone returns the current content and done status
func (s *StreamingSession) GetContentAndDone() (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Read from channel instead of returning full content
	var newContent string
	select {
	case content := <-s.ContentChannel:
		newContent = content
	default:
		// No new content available
	}

	return newContent, s.Done
}

// GetNewContent returns only new content since last call
func (s *StreamingSession) GetNewContent() (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var newContent string

	// Read all available content from channel
	for {
		select {
		case content := <-s.ContentChannel:
			newContent += content
		default:
			// No more content available
			goto done
		}
	}

done:
	return newContent, s.Done
}

// MarkDone marks the session as completed
func (s *StreamingSession) MarkDone() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Done {
		return // Already done
	}

	s.Done = true

	// Close the content channel
	close(s.ContentChannel)

	// Close all subscriber channels
	for _, subscriber := range s.subscribers {
		close(subscriber)
	}
	s.subscribers = nil
}

// IsExpired checks if the session has expired
func (s *StreamingSession) IsExpired() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return time.Now().After(s.ExpiresAt)
}

// Subscribe adds a new subscriber channel
func (s *StreamingSession) Subscribe() <-chan string {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Done {
		// Return a closed channel if already done
		ch := make(chan string)
		close(ch)
		return ch
	}

	ch := make(chan string, 10) // Buffered subscriber channel
	s.subscribers = append(s.subscribers, ch)
	return ch
}

// Unsubscribe removes a subscriber channel
func (s *StreamingSession) Unsubscribe(ch <-chan string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, subscriber := range s.subscribers {
		if subscriber == ch {
			// Close the channel
			close(subscriber)
			// Remove from slice
			s.subscribers = append(s.subscribers[:i], s.subscribers[i+1:]...)
			break
		}
	}
}

// UpdateLastPoll updates the last poll time
func (s *StreamingSession) UpdateLastPoll() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.LastPoll = time.Now()
}

// GetStats returns session statistics
func (s *StreamingSession) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return map[string]interface{}{
		"id":               s.ID,
		"created_at":       s.CreatedAt,
		"last_poll":        s.LastPoll,
		"expires_at":       s.ExpiresAt,
		"done":             s.Done,
		"content_length":   len(s.Content),
		"subscribers":      len(s.subscribers),
		"channel_capacity": cap(s.ContentChannel),
		"channel_length":   len(s.ContentChannel),
	}
}

// Helper function to generate session IDs
func generateSessionID() string {
	return time.Now().Format("20060102150405") + "_" + generateRandomString(8)
}

// Helper function to generate random strings
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[time.Now().Nanosecond()%len(charset)]
	}
	return string(result)
}

func (s *UserSession) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

func (s *UserSession) UpdateLastUsed() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.LastUsed = time.Now()
}

func (s *UserSession) GetID() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.ID
}

func (s *UserSession) GetAccountID() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.AccountID
}
