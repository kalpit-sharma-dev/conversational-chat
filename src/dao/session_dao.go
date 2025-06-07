package dao

import (
	"sync"

	"github.com/banking/ai-agents-banking/src/models"
)

type SessionDAO struct {
	userSessions        map[string]*models.UserSession
	tokenToSession      map[string]string
	streamingSessions   map[string]*models.StreamingSession
	userSessionsMu      sync.RWMutex
	tokenToSessionMu    sync.RWMutex
	streamingSessionsMu sync.RWMutex
}

func NewSessionDAO() *SessionDAO {
	return &SessionDAO{
		userSessions:      make(map[string]*models.UserSession),
		tokenToSession:    make(map[string]string),
		streamingSessions: make(map[string]*models.StreamingSession),
	}
}

func (d *SessionDAO) CreateUserSession(session *models.UserSession) error {
	d.userSessionsMu.Lock()
	defer d.userSessionsMu.Unlock()

	d.userSessions[session.ID] = session

	d.tokenToSessionMu.Lock()
	d.tokenToSession[session.Token] = session.ID
	d.tokenToSessionMu.Unlock()

	return nil
}

func (d *SessionDAO) GetUserSessionByToken(token string) (*models.UserSession, bool) {
	d.tokenToSessionMu.RLock()
	sessionID, exists := d.tokenToSession[token]
	d.tokenToSessionMu.RUnlock()

	if !exists {
		return nil, false
	}

	d.userSessionsMu.RLock()
	session, exists := d.userSessions[sessionID]
	d.userSessionsMu.RUnlock()

	return session, exists
}

func (d *SessionDAO) UpdateUserSession(session *models.UserSession) error {
	d.userSessionsMu.Lock()
	defer d.userSessionsMu.Unlock()

	d.userSessions[session.ID] = session
	return nil
}

func (d *SessionDAO) DeleteUserSession(sessionID, token string) {
	d.userSessionsMu.Lock()
	delete(d.userSessions, sessionID)
	d.userSessionsMu.Unlock()

	d.tokenToSessionMu.Lock()
	delete(d.tokenToSession, token)
	d.tokenToSessionMu.Unlock()
}

func (d *SessionDAO) GetAllUserSessions() map[string]*models.UserSession {
	d.userSessionsMu.RLock()
	defer d.userSessionsMu.RUnlock()

	result := make(map[string]*models.UserSession)
	for k, v := range d.userSessions {
		result[k] = v
	}
	return result
}

func (d *SessionDAO) CreateStreamingSession(session *models.StreamingSession) error {
	d.streamingSessionsMu.Lock()
	defer d.streamingSessionsMu.Unlock()

	d.streamingSessions[session.ID] = session
	return nil
}

func (d *SessionDAO) GetStreamingSession(sessionID string) (*models.StreamingSession, bool) {
	d.streamingSessionsMu.RLock()
	defer d.streamingSessionsMu.RUnlock()

	session, exists := d.streamingSessions[sessionID]
	return session, exists
}

func (d *SessionDAO) DeleteStreamingSession(sessionID string) {
	d.streamingSessionsMu.Lock()
	defer d.streamingSessionsMu.Unlock()

	delete(d.streamingSessions, sessionID)
}

func (d *SessionDAO) GetAllStreamingSessions() map[string]*models.StreamingSession {
	d.streamingSessionsMu.RLock()
	defer d.streamingSessionsMu.RUnlock()

	result := make(map[string]*models.StreamingSession)
	for k, v := range d.streamingSessions {
		result[k] = v
	}
	return result
}
