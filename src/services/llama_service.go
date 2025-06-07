package services

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/banking/ai-agents-banking/src/models"
)

type LlamaService struct {
	baseURL    string
	httpClient *http.Client
	model      string
	apiKey     string
	apiURL     string
}

type LlamaRequest struct {
	Model   string                 `json:"model"`
	Prompt  string                 `json:"prompt"`
	Stream  bool                   `json:"stream"`
	Options map[string]interface{} `json:"options,omitempty"`
}

type LlamaResponse struct {
	Model              string `json:"model"`
	CreatedAt          string `json:"created_at"`
	Response           string `json:"response"`
	Done               bool   `json:"done"`
	DoneReason         string `json:"done_reason,omitempty"`
	Context            []int  `json:"context,omitempty"`
	TotalDuration      int64  `json:"total_duration,omitempty"`
	LoadDuration       int64  `json:"load_duration,omitempty"`
	PromptEvalCount    int    `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int64  `json:"prompt_eval_duration,omitempty"`
	EvalCount          int    `json:"eval_count,omitempty"`
	EvalDuration       int64  `json:"eval_duration,omitempty"`
}

func NewLlamaService(baseURL string) *LlamaService {
	return &LlamaService{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Minute, // Increase timeout for longer responses
		},
		model:  "llama3", // Updated to use llama3
		apiURL: baseURL,
	}
}

func (s *LlamaService) GenerateResponse(message string, agent *models.AgentResponse) (string, error) {
	// Create a new streaming session
	session := &models.StreamingSession{
		ID:        fmt.Sprintf("session_%d", time.Now().UnixNano()),
		CreatedAt: time.Now(),
		Content:   "",
		Done:      false,
	}

	// Build the prompt
	prompt := s.BuildPromptWithContext(&models.StreamingContext{
		Message: message,
		Intent: &models.Intent{
			Name:       agent.AgentName,
			Confidence: 1.0,
			Entities:   make(map[string]interface{}),
		},
	})

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Start streaming
	s.QueryStreamingWithContext(ctx, prompt, session)

	// Wait for completion or timeout
	for !session.Done {
		select {
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				return "", fmt.Errorf("request timed out after 2 minutes")
			}
			return "", ctx.Err()
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}

	// Return the final content
	return session.Content, nil
}

func (s *LlamaService) QueryStreamingWithContext(ctx context.Context, prompt string, session *models.StreamingSession) {
	log.Printf("Starting Llama streaming request with prompt length: %d", len(prompt))

	// Prepare the request with banking-specific options
	requestBody := LlamaRequest{
		Model:  s.model,
		Prompt: prompt,
		Stream: true,
		Options: map[string]interface{}{
			"temperature":    0.7,                         // Balanced creativity
			"top_p":          0.9,                         // Good diversity
			"top_k":          40,                          // Reasonable token selection
			"max_tokens":     1000,                        // Reasonable response length
			"stop":           []string{"Human:", "User:"}, // Stop tokens
			"repeat_penalty": 1.1,                         // Avoid repetition
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		log.Printf("Error marshaling request body: %v", err)
		session.AppendContent(fmt.Sprintf("Error: Failed to prepare request - %v", err))
		session.MarkDone()
		return
	}

	log.Printf("Request body: %s", string(jsonBody))

	// Create the request with context
	req, err := http.NewRequestWithContext(ctx, "POST", s.apiURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		session.AppendContent(fmt.Sprintf("Error: Failed to create request - %v", err))
		session.MarkDone()
		return
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if s.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+s.apiKey)
	}

	log.Printf("Sending request to: %s", s.apiURL)

	// Send the request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		log.Printf("Error sending request: %v", err)
		if strings.Contains(err.Error(), "connection refused") {
			session.AppendContent("Error: Cannot connect to Llama service. Please ensure the Llama server is running.")
		} else {
			session.AppendContent(fmt.Sprintf("Error: Network error - %v", err))
		}
		session.MarkDone()
		return
	}
	defer resp.Body.Close()

	log.Printf("Response status: %s", resp.Status)

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Error from Llama API: %s - %s", resp.Status, string(body))

		if resp.StatusCode == 404 {
			session.AppendContent("Error: Llama model not found. Please ensure the model is downloaded and available.")
		} else if resp.StatusCode == 500 {
			session.AppendContent("Error: Llama server error. Please check the server logs.")
		} else {
			session.AppendContent(fmt.Sprintf("Error: API returned status %s", resp.Status))
		}
		session.MarkDone()
		return
	}

	// Process the streaming response
	s.processStreamingResponse(ctx, session, resp)
}

func (s *LlamaService) processStreamingResponse(ctx context.Context, session *models.StreamingSession, resp *http.Response) error {
	defer resp.Body.Close()

	reader := bufio.NewReader(resp.Body)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					return nil
				}
				return fmt.Errorf("error reading response: %v", err)
			}

			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			// Parse the JSON response
			var llamaResp struct {
				Response string `json:"response"`
				Done     bool   `json:"done"`
			}

			if err := json.Unmarshal([]byte(line), &llamaResp); err != nil {
				log.Printf("Error parsing Llama response: %v", err)
				continue
			}

			// If we have a response, append it to the session content
			if llamaResp.Response != "" {
				// Split the response into words for more granular streaming
				words := strings.Fields(llamaResp.Response)
				for _, word := range words {
					// Add each word with a small delay
					session.AppendContent(word + " ")
					time.Sleep(50 * time.Millisecond) // Small delay between words
				}
			}

			// If the response is done, mark the session as complete
			if llamaResp.Done {
				session.MarkDone()
				return nil
			}
		}
	}
}

func (s *LlamaService) cleanContent(content string) string {
	// Remove any unwanted characters or formatting
	content = strings.TrimSpace(content)

	// Remove any control characters but preserve newlines
	var cleaned strings.Builder
	for _, r := range content {
		if r >= 32 || r == '\n' || r == '\t' {
			cleaned.WriteRune(r)
		}
	}

	return cleaned.String()
}

// BuildPromptWithContext builds a prompt with the given context
func (s *LlamaService) BuildPromptWithContext(ctx *models.StreamingContext) string {
	var promptBuilder strings.Builder

	// Add system context
	promptBuilder.WriteString("You are a helpful AI banking assistant. ")
	promptBuilder.WriteString("You help customers with banking operations like transfers, balance checks, adding payees, loans, and general banking questions. ")
	promptBuilder.WriteString("Be concise, helpful, and professional in your responses.\n\n")

	// Add conversation history if available
	if ctx.Conversation != nil && len(ctx.Conversation.Messages) > 0 {
		promptBuilder.WriteString("Conversation history:\n")
		// Include last 3 exchanges for context
		start := len(ctx.Conversation.Messages) - 6
		if start < 0 {
			start = 0
		}

		for i := start; i < len(ctx.Conversation.Messages); i++ {
			msg := ctx.Conversation.Messages[i]
			role := "Human"
			if msg.Role == "assistant" {
				role = "Assistant"
			}
			promptBuilder.WriteString(fmt.Sprintf("%s: %s\n", role, msg.Content))
		}
		promptBuilder.WriteString("\n")
	}

	// Add intent information
	if ctx.Intent != nil && ctx.Intent.Name != "general" {
		promptBuilder.WriteString(fmt.Sprintf("Intent: %s\n", ctx.Intent.Name))
		if len(ctx.Intent.Entities) > 0 {
			promptBuilder.WriteString("Entities: ")
			for key, value := range ctx.Intent.Entities {
				promptBuilder.WriteString(fmt.Sprintf("%s=%v ", key, value))
			}
			promptBuilder.WriteString("\n")
		}
		promptBuilder.WriteString("\n")
	}

	// Add current message
	promptBuilder.WriteString(fmt.Sprintf("Human: %s\n", ctx.Message))
	promptBuilder.WriteString("Assistant:")

	return promptBuilder.String()
}

// QueryStreaming is a backward compatibility wrapper
func (s *LlamaService) QueryStreaming(message string, session *models.StreamingSession) {
	ctx := &models.StreamingContext{
		Message: message,
	}
	prompt := s.BuildPromptWithContext(ctx)
	s.QueryStreamingWithContext(context.Background(), prompt, session)
}
