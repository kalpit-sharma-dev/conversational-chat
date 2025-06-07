package services

import (
	"sync"

	"github.com/banking/ai-agents-banking/src/models"
)

type ConversationService struct {
	mu            sync.RWMutex
	conversations map[string]*models.Conversation
	memories      map[string]*ConversationMemory
}

func NewConversationService() *ConversationService {
	return &ConversationService{
		conversations: make(map[string]*models.Conversation),
		memories:      make(map[string]*ConversationMemory),
	}
}

func (s *ConversationService) GetOrCreateConversation(sessionID string) *models.Conversation {
	s.mu.Lock()
	defer s.mu.Unlock()

	conv, exists := s.conversations[sessionID]
	if !exists {
		conv = &models.Conversation{
			ID:       sessionID,
			Messages: make([]models.Message, 0),
			Context:  make(map[string]interface{}),
		}
		s.conversations[sessionID] = conv
		s.memories[sessionID] = NewConversationMemory()
	}

	return conv
}

func (s *ConversationService) AddMessage(sessionID, role, content, intent string, actions []string, entities map[string]interface{}, agentName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	conv, exists := s.conversations[sessionID]
	if !exists {
		conv = s.GetOrCreateConversation(sessionID)
	}

	msg := models.Message{
		Role:      role,
		Content:   content,
		Intent:    intent,
		Actions:   actions,
		Entities:  entities,
		AgentName: agentName,
	}

	conv.Messages = append(conv.Messages, msg)

	// Update conversation memory
	if memory, exists := s.memories[sessionID]; exists {
		memory.AddMessage(role, content, intent, entities)
	}

	return nil
}

func (s *ConversationService) GetConversationHistory(sessionID string) []models.Message {
	s.mu.RLock()
	defer s.mu.RUnlock()

	conv, exists := s.conversations[sessionID]
	if !exists {
		return nil
	}

	return conv.Messages
}

func (s *ConversationService) GetConversationMemory(sessionID string) *ConversationMemory {
	s.mu.RLock()
	defer s.mu.RUnlock()

	memory, exists := s.memories[sessionID]
	if !exists {
		return nil
	}

	return memory
}

func (s *ConversationService) ClearConversation(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.conversations, sessionID)
	if memory, exists := s.memories[sessionID]; exists {
		memory.Clear()
	}
}

func (s *ConversationService) UpdateContext(sessionID string, key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	conv, exists := s.conversations[sessionID]
	if !exists {
		conv = s.GetOrCreateConversation(sessionID)
	}

	conv.Context[key] = value

	// Update memory context
	if memory, exists := s.memories[sessionID]; exists {
		memory.SetContext(key, value)
	}
}

func (s *ConversationService) GetContext(sessionID string, key string) (interface{}, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	conv, exists := s.conversations[sessionID]
	if !exists {
		return nil, false
	}

	value, exists := conv.Context[key]
	return value, exists
}

// DeleteConversation completely removes a conversation
func (s *ConversationService) DeleteConversation(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.conversations, sessionID)
	if memory, exists := s.memories[sessionID]; exists {
		memory.Clear()
	}
}

// GetAllConversations returns all active conversations (for admin/debugging)
func (s *ConversationService) GetAllConversations() map[string]*models.Conversation {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]*models.Conversation)
	for k, v := range s.conversations {
		result[k] = v
	}
	return result
}

func (s *ConversationService) GetActiveConversationsCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.conversations)
}
