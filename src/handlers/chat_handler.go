package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/banking/ai-agents-banking/src/models"
	"github.com/banking/ai-agents-banking/src/services"
	"github.com/banking/ai-agents-banking/src/utils"
)

type ChatHandler struct {
	sessionService      *services.SessionService
	agentService        *services.AgentService
	conversationService *services.ConversationService
	intentService       *services.IntentRecognitionService
	llamaService        *services.LlamaService
	toolRegistry        *services.ToolRegistry
}

func NewChatHandler(
	sessionService *services.SessionService,
	agentService *services.AgentService,
	conversationService *services.ConversationService,
	intentService *services.IntentRecognitionService,
	llamaService *services.LlamaService,
	toolRegistry *services.ToolRegistry,
) *ChatHandler {
	return &ChatHandler{
		sessionService:      sessionService,
		agentService:        agentService,
		conversationService: conversationService,
		intentService:       intentService,
		llamaService:        llamaService,
		toolRegistry:        toolRegistry,
	}
}

func (h *ChatHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("=== Chat request received from %s ===", r.RemoteAddr)
	log.Printf("Request headers: %v", r.Header)
	log.Printf("Request method: %s", r.Method)
	log.Printf("Request URL: %s", r.URL.String())

	// Handle CORS preflight
	if r.Method == http.MethodOptions {
		h.setCORSHeaders(w)
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		log.Printf("Invalid method: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Set proper SSE headers first
	h.setSSEHeaders(w)

	// Get flusher after setting headers
	flusher, ok := w.(http.Flusher)
	if !ok {
		log.Printf("Streaming not supported by client")
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Get session from context
	session, ok := utils.GetUserSessionFromContext(r)
	if !ok {
		log.Printf("Session not found in context")
		http.Error(w, "Session not found", http.StatusUnauthorized)
		return
	}

	// Parse request
	var req models.ChatRequest
	if r.Method == http.MethodGet {
		req = models.ChatRequest{
			Message:   r.URL.Query().Get("message"),
			Token:     r.URL.Query().Get("token"),
			SessionID: r.URL.Query().Get("session_id"),
			Stream:    r.URL.Query().Get("stream") == "true",
		}
		log.Printf("[Request] Parsed GET parameters - Message: %s, Token: %s, SessionID: %s, Stream: %v",
			req.Message, req.Token, req.SessionID, req.Stream)
	} else {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("Error decoding request body: %v", err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
	}

	if req.Message == "" {
		http.Error(w, "Message is required", http.StatusBadRequest)
		return
	}

	log.Printf("[Session] Found session: %+v", session)
	conversation := h.conversationService.GetOrCreateConversation(session.ID)
	log.Printf("[Conversation] Using conversation: %+v", conversation)

	// Create a unique streaming session ID
	streamingSessionID := fmt.Sprintf("%s_%d", session.ID, time.Now().UnixNano())

	// Create a streaming session
	streamingSession := models.NewStreamingSession(session.Token)
	if streamingSession == nil {
		log.Printf("Failed to create streaming session")
		http.Error(w, "Failed to create streaming session", http.StatusInternalServerError)
		return
	}
	streamingSession.ID = streamingSessionID

	// Save streaming session
	h.sessionService.SaveStreamingSession(streamingSession)

	// Send initial status
	if err := h.writeSSEData(w, fmt.Sprintf(`{"session_id": "%s", "status": "processing"}`, streamingSessionID)); err != nil {
		log.Printf("Failed to send initial status: %v", err)
		return
	}
	flusher.Flush()

	// Create a channel to track completion
	done := make(chan bool)

	// Process the message in a goroutine
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Recovered from panic in message processing: %v", r)
				h.writeSSEError(w, "Internal server error during message processing")
			}
			close(done)
		}()

		// Detect intent
		intent := h.intentService.RecognizeIntent(req.Message)
		log.Printf("[Intent] Detected intent: %+v", intent)

		// Add user message to conversation
		err := h.conversationService.AddMessage(session.ID, "user", req.Message, intent.Name, nil, intent.Entities, "")
		if err != nil {
			log.Printf("[Message] Error adding user message: %v", err)
			h.writeSSEError(w, fmt.Sprintf("Error adding message: %v", err))
			return
		}

		log.Printf("[Message] Received message from session %s: %s (Intent: %s, Confidence: %.2f)\n",
			session.ID, req.Message, intent.Name, intent.Confidence)

		// Build enhanced prompt with banking context
		prompt := h.buildBankingPrompt(req.Message, conversation, intent)
		if prompt == "" {
			log.Printf("Failed to build prompt")
			h.writeSSEError(w, "Failed to build prompt")
			return
		}
		log.Printf("[Stream] Built prompt: %s", prompt)

		// Query the Llama model with the request context
		h.llamaService.QueryStreamingWithContext(context.Background(), prompt, streamingSession)
	}()

	// Stream the response
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	var lastContent string
	for {
		select {
		case <-done:
			// Send final message
			if err := h.writeSSEData(w, `{"response": "", "done": true}`); err != nil {
				log.Printf("Error sending final message: %v", err)
			}
			flusher.Flush()
			return
		case <-ticker.C:
			content, isDone := streamingSession.GetContentAndDone()
			if content != lastContent {
				// Only send if content has changed
				if err := h.writeSSEData(w, fmt.Sprintf(`{"response": "%s", "done": false}`, content)); err != nil {
					log.Printf("Error sending content update: %v", err)
					return
				}
				flusher.Flush()
				lastContent = content
			}
			if isDone {
				// Send final message
				if err := h.writeSSEData(w, `{"response": "", "done": true}`); err != nil {
					log.Printf("Error sending final message: %v", err)
				}
				flusher.Flush()
				return
			}
		}
	}
}

// splitIntoChunks splits a string into chunks of specified size
func splitIntoChunks(s string, chunkSize int) []string {
	var chunks []string
	for i := 0; i < len(s); i += chunkSize {
		end := i + chunkSize
		if end > len(s) {
			end = len(s)
		}
		chunks = append(chunks, s[i:end])
	}
	return chunks
}

// handleStreamRequest handles SSE streaming for chat responses
func (h *ChatHandler) handleStreamRequest(w http.ResponseWriter, r *http.Request) {
	log.Printf("=== Stream request received from %s ===", r.RemoteAddr)

	// Set proper SSE headers
	h.setSSEHeaders(w)

	// Get session from context
	session, ok := utils.GetUserSessionFromContext(r)
	if !ok {
		log.Printf("Session not found in context")
		h.writeSSEError(w, "Session not found")
		return
	}

	// Get session ID from query parameters
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		sessionID = session.ID
	}

	log.Printf("[Stream] Request for session ID: %s", sessionID)

	// Get streaming session
	streamingSession, exists := h.sessionService.GetStreamingSession(sessionID)
	if !exists {
		log.Printf("[Stream] Streaming session not found: %s", sessionID)
		h.writeSSEError(w, "Streaming session not found")
		return
	}

	// Send initial connection message
	h.writeSSEEvent(w, "open", "Connection established")
	flusher, ok := w.(http.Flusher)
	if ok {
		flusher.Flush()
	}

	// Subscribe to the streaming session
	contentChan := streamingSession.Subscribe()

	// Create a ticker for keep-alive messages
	keepAliveTicker := time.NewTicker(30 * time.Second)
	defer keepAliveTicker.Stop()

	// Create a done channel
	done := make(chan struct{})
	defer close(done)

	// Check if the session is already done
	if streamingSession.Done {
		// Send the full content and done message
		h.writeSSEEvent(w, "message", streamingSession.Content)
		h.writeSSEEvent(w, "done", "[DONE]")

		flusher.Flush()

		return
	}

	// Start streaming in a goroutine
	go func() {
		defer streamingSession.Unsubscribe(contentChan)

		for {
			select {
			case content, ok := <-contentChan:
				if !ok {
					// Channel closed, session is done
					h.writeSSEEvent(w, "done", "[DONE]")

					flusher.Flush()
					return
				}

				if content != "" {
					// Send the content chunk
					h.writeSSEEvent(w, "message", content)

					flusher.Flush()

				}

			case <-keepAliveTicker.C:
				// Send keep-alive message
				w.Write([]byte(": keep-alive\n\n"))

				flusher.Flush()

			case <-r.Context().Done():
				// Client disconnected
				log.Printf("[Stream] Client disconnected")
				return

			case <-done:
				return
			}
		}
	}()

	// Wait for the request context to be done (client disconnects)
	<-r.Context().Done()
}

// This method is no longer used - replaced by handleStreamRequest
func (h *ChatHandler) processMessageStream(w http.ResponseWriter, r *http.Request, req models.ChatRequest, session *models.UserSession) {
	log.Printf("[DEPRECATED] processMessageStream called - this method is no longer used")
}

func (h *ChatHandler) buildBankingPrompt(message string, conversation *models.Conversation, intent *services.Intent) string {
	// Start with system context
	prompt := "You are a helpful banking assistant. Provide clear, concise responses about banking services.\n\n"

	// Add conversation history if available
	if conversation != nil && len(conversation.Messages) > 0 {
		prompt += "Previous conversation:\n"
		for _, msg := range conversation.Messages {
			prompt += fmt.Sprintf("%s: %s\n", msg.Role, msg.Content)
		}
		prompt += "\n"
	}

	// Add current message with intent context
	prompt += fmt.Sprintf("User (Intent: %s): %s\nAssistant:", intent.Name, message)
	return prompt
}

func (h *ChatHandler) setSSEHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("X-Accel-Buffering", "no") // Disable proxy buffering
	h.setCORSHeaders(w)
}

func (h *ChatHandler) setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Max-Age", "86400")
}

func (h *ChatHandler) writeSSEData(w http.ResponseWriter, data string) error {
	_, err := fmt.Fprintf(w, "data: %s\n\n", data)
	return err
}

func (h *ChatHandler) writeSSEError(w http.ResponseWriter, errorMsg string) {
	fmt.Fprintf(w, "data: {\"error\":\"%s\"}\n\n", errorMsg)
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (h *ChatHandler) writeSSEEvent(w http.ResponseWriter, event string, data string) {
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, data)
}

// Keep existing Stream and Poll methods for backward compatibility
func (h *ChatHandler) Stream(w http.ResponseWriter, r *http.Request) {
	// Redirect to main ServeHTTP for consistency
	h.ServeHTTP(w, r)
}

func (h *ChatHandler) Poll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/chat/poll/")
	sessionID := path

	// Get session from context for validation
	_, ok := utils.GetUserSessionFromContext(r)
	if !ok {
		http.Error(w, "Session not found in context", http.StatusUnauthorized)
		return
	}

	session, exists := h.sessionService.GetStreamingSession(sessionID)
	if !exists {
		http.Error(w, "Streaming session not found", http.StatusNotFound)
		return
	}

	content, done := session.GetContentAndDone()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"content": content,
		"done":    done,
	})
}
