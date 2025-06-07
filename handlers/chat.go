package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/banking/ai-agents-banking/services"
)

// ChatHandler handles chat-related HTTP requests
type ChatHandler struct {
	sessionService      services.SessionService
	agentService        services.AgentService
	conversationService services.ConversationService
	intentService       services.IntentRecognitionService
	llamaService        services.LlamaService
	toolRegistry        services.ToolRegistry
}

// NewChatHandler creates a new ChatHandler instance
func NewChatHandler(
	sessionService services.SessionService,
	agentService services.AgentService,
	conversationService services.ConversationService,
	intentService services.IntentRecognitionService,
	llamaService services.LlamaService,
	toolRegistry services.ToolRegistry,
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

// HandleChat processes chat requests and streams responses
func (h *ChatHandler) HandleChat(w http.ResponseWriter, r *http.Request) {
	// Set headers for streaming
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Transfer-Encoding", "chunked")

	// Get session ID from query parameter
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		http.Error(w, "Session ID is required", http.StatusBadRequest)
		return
	}

	// Get session from service
	session, err := h.sessionService.GetSession(sessionID)
	if err != nil {
		http.Error(w, "Invalid session", http.StatusUnauthorized)
		return
	}

	// Create a flusher
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Send initial status message
	statusMsg := map[string]string{
		"session_id": sessionID,
		"status":     "processing",
	}
	statusJSON, _ := json.Marshal(statusMsg)
	fmt.Fprintf(w, "data: %s\n\n", string(statusJSON))
	flusher.Flush()

	// Get the message from the request body
	var requestBody struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		sendError(w, flusher, fmt.Errorf("error decoding request: %v", err))
		return
	}

	// Process the message
	response, err := h.processMessage(requestBody.Message, session)
	if err != nil {
		sendError(w, flusher, err)
		return
	}

	// Stream the response
	for _, chunk := range response {
		// Create SSE message
		msg := map[string]interface{}{
			"response": chunk,
			"done":     false,
		}
		msgJSON, _ := json.Marshal(msg)
		fmt.Fprintf(w, "data: %s\n\n", string(msgJSON))
		flusher.Flush()
		time.Sleep(50 * time.Millisecond) // Adjust delay as needed
	}

	// Send completion message
	doneMsg := map[string]interface{}{
		"response": "",
		"done":     true,
	}
	doneJSON, _ := json.Marshal(doneMsg)
	fmt.Fprintf(w, "data: %s\n\n", string(doneJSON))
	flusher.Flush()
}

// processMessage processes the chat message using your existing services
func (h *ChatHandler) processMessage(message string, session *services.Session) ([]string, error) {
	// 1. Recognize intent
	intent, err := h.intentService.RecognizeIntent(message)
	if err != nil {
		return nil, fmt.Errorf("error recognizing intent: %v", err)
	}

	// 2. Get appropriate agent
	agent, err := h.agentService.GetAgentForIntent(intent)
	if err != nil {
		return nil, fmt.Errorf("error getting agent: %v", err)
	}

	// 3. Process with LLM
	response, err := h.llamaService.GenerateResponse(message, agent)
	if err != nil {
		return nil, fmt.Errorf("error generating response: %v", err)
	}

	// 4. Split response into chunks for streaming
	chunks := splitIntoChunks(response)

	// 5. Update conversation history
	err = h.conversationService.AddMessage(session.ID, message, response)
	if err != nil {
		return nil, fmt.Errorf("error updating conversation: %v", err)
	}

	return chunks, nil
}

// splitIntoChunks splits a response into smaller chunks for streaming
func splitIntoChunks(response string) []string {
	// Split by sentences and words for a more natural streaming effect
	words := strings.Fields(response)
	chunks := make([]string, 0)
	currentChunk := ""

	for _, word := range words {
		currentChunk += word + " "
		if strings.HasSuffix(word, ".") || strings.HasSuffix(word, "!") || strings.HasSuffix(word, "?") {
			chunks = append(chunks, strings.TrimSpace(currentChunk))
			currentChunk = ""
		}
	}

	// Add any remaining text
	if currentChunk != "" {
		chunks = append(chunks, strings.TrimSpace(currentChunk))
	}

	return chunks
}

// sendError sends an error message in SSE format
func sendError(w http.ResponseWriter, flusher http.Flusher, err error) {
	errorMsg := map[string]string{
		"error": err.Error(),
	}
	errorJSON, _ := json.Marshal(errorMsg)
	fmt.Fprintf(w, "data: %s\n\n", string(errorJSON))
	flusher.Flush()
}
