package services

import (
	"log"
	"sync"
	"time"

	"github.com/banking/ai-agents-banking/src/dao"
	"github.com/banking/ai-agents-banking/src/models"
	"github.com/banking/ai-agents-banking/src/utils"
)

// SessionService implements the SessionValidator interface to avoid circular imports
type SessionService struct {
	sessionDAO *dao.SessionDAO
	sessions   map[string]*models.StreamingSession
	mu         sync.RWMutex
}

func NewSessionService(sessionDAO *dao.SessionDAO) *SessionService {
	return &SessionService{
		sessionDAO: sessionDAO,
		sessions:   make(map[string]*models.StreamingSession),
	}
}

func (s *SessionService) CreateUserSession(accountID string) (*models.UserSession, error) {
	sessionID := utils.GenerateSessionID()
	token := utils.GenerateToken()
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour) // TokenExpiry

	session := &models.UserSession{
		ID:        sessionID,
		Token:     token,
		CreatedAt: now,
		ExpiresAt: expiresAt,
		LastUsed:  now,
		AccountID: accountID,
	}

	err := s.sessionDAO.CreateUserSession(session)
	if err != nil {
		return nil, err
	}

	log.Printf("Created new session: %s with token: %s", sessionID, token[:8]+"...")
	return session, nil
}

// ValidateToken implements the SessionValidator interface
// Returns interface{} to avoid circular import issues
func (s *SessionService) ValidateToken(token string) (interface{}, bool) {
	if token == "" {
		return nil, false
	}

	session, exists := s.sessionDAO.GetUserSessionByToken(token)
	if !exists {
		return nil, false
	}

	if session.IsExpired() {
		s.sessionDAO.DeleteUserSession(session.ID, session.Token)
		return nil, false
	}

	session.UpdateLastUsed()
	s.sessionDAO.UpdateUserSession(session)
	return session, true
}

// ValidateTokenTyped returns the typed session for internal use
func (s *SessionService) ValidateTokenTyped(token string) (*models.UserSession, bool) {
	session, valid := s.ValidateToken(token)
	if !valid {
		return nil, false
	}

	if typedSession, ok := session.(*models.UserSession); ok {
		return typedSession, true
	}

	return nil, false
}

func (s *SessionService) CreateStreamingSession(token string) (*models.StreamingSession, error) {
	sessionID := utils.GenerateSessionID()
	now := time.Now()

	// Create new streaming session with proper initialization
	session := models.NewStreamingSession(token)
	session.ID = sessionID
	session.LastPoll = now
	session.CreatedAt = now
	session.ExpiresAt = now.Add(30 * time.Minute) // SessionExpiry

	s.mu.Lock()
	s.sessions[sessionID] = session
	s.mu.Unlock()

	log.Printf("Created new streaming session: %s", sessionID)
	return session, nil
}

func (s *SessionService) GetStreamingSession(sessionID string) (*models.StreamingSession, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, exists := s.sessions[sessionID]
	return session, exists
}

func (s *SessionService) SaveStreamingSession(session *models.StreamingSession) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions[session.ID] = session
	log.Printf("Saved streaming session: %s", session.ID)
}

func (s *SessionService) StartCleanupRoutine() {
	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()

		// Clean up expired user sessions
		sessions := s.sessionDAO.GetAllUserSessions()
		for id, session := range sessions {
			if session.IsExpired() {
				s.sessionDAO.DeleteUserSession(id, session.Token)
				log.Printf("Cleaned up expired user session %s", id)
			}
		}

		// Clean up expired streaming sessions
		streamingSessions := s.sessionDAO.GetAllStreamingSessions()
		for id, session := range streamingSessions {
			if session.IsExpired() || (session.Done && now.Sub(session.LastPoll) > 5*time.Minute) {
				s.sessionDAO.DeleteStreamingSession(id)
				log.Printf("Cleaned up streaming session %s", id)
			}
		}
	}
}

func (s *SessionService) GetActiveSessionsCount() int {
	sessions := s.sessionDAO.GetAllUserSessions()
	count := 0
	now := time.Now()

	for _, session := range sessions {
		if !session.IsExpired() && now.Before(session.ExpiresAt) {
			count++
		}
	}

	return count
}

func (s *SessionService) CleanupExpiredSessions() {
	s.mu.Lock()
	defer s.mu.Unlock()

	//now := time.Now()
	for id, session := range s.sessions {
		if session.IsExpired() {
			close(session.ContentChannel) // Close the channel
			delete(s.sessions, id)
			log.Printf("Cleaned up expired session: %s", id)
		}
	}
}
