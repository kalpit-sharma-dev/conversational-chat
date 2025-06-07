package services

import (
	"regexp"
	"strings"
)

type Intent struct {
	Name       string
	Confidence float64
	Entities   map[string]interface{}
	Patterns   []*regexp.Regexp
	Keywords   map[string]float64
}

type IntentRecognitionService struct {
	intents map[string]*Intent
}

func NewIntentRecognitionService() *IntentRecognitionService {
	service := &IntentRecognitionService{
		intents: make(map[string]*Intent),
	}
	service.initializeBankingIntents()
	return service
}

func (s *IntentRecognitionService) initializeBankingIntents() {
	// Fund Transfer Intent
	s.intents["fund_transfer"] = &Intent{
		Name: "fund_transfer",
		Patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)transfer\s+(\d+(?:\.\d{2})?)\s+(?:to|for)\s+([a-zA-Z0-9@._-]+)`),
			regexp.MustCompile(`(?i)send\s+(\d+(?:\.\d{2})?)\s+(?:to|for)\s+([a-zA-Z0-9@._-]+)`),
		},
		Keywords: map[string]float64{
			"transfer": 1.0,
			"send":     0.9,
			"money":    0.8,
			"amount":   0.7,
			"upi":      0.6,
			"neft":     0.6,
			"imps":     0.6,
		},
	}

	// Balance Check Intent
	s.intents["check_balance"] = &Intent{
		Name: "check_balance",
		Patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)balance\s+(?:of|for|in)\s+account\s+(\d+)`),
			regexp.MustCompile(`(?i)how\s+much\s+(?:do\s+I\s+have|is\s+in\s+my\s+account)`),
		},
		Keywords: map[string]float64{
			"balance":   1.0,
			"amount":    0.8,
			"available": 0.7,
			"check":     0.6,
		},
	}

	// Add Payee Intent
	s.intents["add_payee"] = &Intent{
		Name: "add_payee",
		Patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)add\s+(?:new\s+)?payee\s+(?:named\s+)?([a-zA-Z\s]+)`),
			regexp.MustCompile(`(?i)save\s+(?:new\s+)?beneficiary\s+(?:named\s+)?([a-zA-Z\s]+)`),
		},
		Keywords: map[string]float64{
			"payee":       1.0,
			"beneficiary": 0.9,
			"add":         0.8,
			"save":        0.7,
			"new":         0.6,
		},
	}

	// Fixed Deposit Intent
	s.intents["create_fd"] = &Intent{
		Name: "create_fd",
		Patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)create\s+(?:a\s+)?fixed\s+deposit\s+(?:of\s+)?(\d+(?:\.\d{2})?)\s+(?:for\s+)?(\d+)\s+(?:months|years)`),
			regexp.MustCompile(`(?i)open\s+(?:a\s+)?fd\s+(?:of\s+)?(\d+(?:\.\d{2})?)\s+(?:for\s+)?(\d+)\s+(?:months|years)`),
		},
		Keywords: map[string]float64{
			"fixed":   1.0,
			"deposit": 0.9,
			"fd":      0.8,
			"create":  0.7,
			"open":    0.7,
			"tenure":  0.6,
		},
	}
}

func (s *IntentRecognitionService) RecognizeIntent(message string) *Intent {
	var bestIntent *Intent
	highestConfidence := 0.0

	for _, intent := range s.intents {
		confidence := s.calculateConfidence(message, intent)
		if confidence > highestConfidence {
			highestConfidence = confidence
			bestIntent = intent
		}
	}

	if bestIntent != nil {
		// Create a copy of the intent with extracted entities
		result := &Intent{
			Name:       bestIntent.Name,
			Confidence: highestConfidence,
			Entities:   s.extractEntities(message, bestIntent),
		}
		return result
	}

	// Default to general query if no intent matches
	return &Intent{
		Name:       "general_query",
		Confidence: 0.0,
		Entities:   make(map[string]interface{}),
	}
}

func (s *IntentRecognitionService) calculateConfidence(message string, intent *Intent) float64 {
	confidence := 0.0
	words := strings.Fields(strings.ToLower(message))

	// Check patterns
	for _, pattern := range intent.Patterns {
		if pattern.MatchString(message) {
			confidence += 0.5
		}
	}

	// Check keywords
	for _, word := range words {
		if weight, exists := intent.Keywords[word]; exists {
			confidence += weight
		}
	}

	return confidence
}

func (s *IntentRecognitionService) extractEntities(message string, intent *Intent) map[string]interface{} {
	entities := make(map[string]interface{})

	switch intent.Name {
	case "fund_transfer":
		if matches := regexp.MustCompile(`(\d+(?:\.\d{2})?)\s+(?:to|for)\s+([a-zA-Z0-9@._-]+)`).FindStringSubmatch(message); len(matches) > 2 {
			entities["amount"] = matches[1]
			entities["recipient"] = matches[2]
		}

	case "check_balance":
		if matches := regexp.MustCompile(`account\s+(\d+)`).FindStringSubmatch(message); len(matches) > 1 {
			entities["account_number"] = matches[1]
		}

	case "add_payee":
		if matches := regexp.MustCompile(`(?:named\s+)?([a-zA-Z\s]+)`).FindStringSubmatch(message); len(matches) > 1 {
			entities["payee_name"] = strings.TrimSpace(matches[1])
		}

	case "create_fd":
		if matches := regexp.MustCompile(`(\d+(?:\.\d{2})?)\s+(?:for\s+)?(\d+)\s+(months|years)`).FindStringSubmatch(message); len(matches) > 3 {
			entities["amount"] = matches[1]
			entities["tenure"] = matches[2]
			entities["tenure_unit"] = matches[3]
		}
	}

	return entities
}
