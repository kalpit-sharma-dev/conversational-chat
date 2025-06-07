package services

import (
	"sync"
	"time"
)

type ConversationMemory struct {
	mu       sync.RWMutex
	messages []Message
	entities map[string]interface{}
	context  map[string]interface{}
}

type Message struct {
	Role      string                 `json:"role"`
	Content   string                 `json:"content"`
	Timestamp time.Time              `json:"timestamp"`
	Entities  map[string]interface{} `json:"entities"`
	Intent    string                 `json:"intent"`
}

func NewConversationMemory() *ConversationMemory {
	return &ConversationMemory{
		messages: make([]Message, 0),
		entities: make(map[string]interface{}),
		context:  make(map[string]interface{}),
	}
}

func (cm *ConversationMemory) AddMessage(role, content, intent string, entities map[string]interface{}) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	msg := Message{
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
		Entities:  entities,
		Intent:    intent,
	}

	cm.messages = append(cm.messages, msg)

	// Keep only last 5 messages for context
	if len(cm.messages) > 5 {
		cm.messages = cm.messages[len(cm.messages)-5:]
	}

	// Update entities
	for k, v := range entities {
		cm.entities[k] = v
	}
}

func (cm *ConversationMemory) GetRecentMessages() []Message {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.messages
}

func (cm *ConversationMemory) GetEntities() map[string]interface{} {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.entities
}

func (cm *ConversationMemory) SetContext(key string, value interface{}) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.context[key] = value
}

func (cm *ConversationMemory) GetContext(key string) (interface{}, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	val, exists := cm.context[key]
	return val, exists
}

func (cm *ConversationMemory) Clear() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.messages = make([]Message, 0)
	cm.entities = make(map[string]interface{})
	cm.context = make(map[string]interface{})
}
