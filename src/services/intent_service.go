package services

import (
	"regexp"
	"strings"

	"github.com/banking/ai-agents-banking/src/models"
)

type IntentService struct {
	patterns []models.IntentPattern
}

func NewIntentService() *IntentService {
	service := &IntentService{}
	service.initializeIntentPatterns()
	return service
}

func (s *IntentService) RecognizeIntent(message string) models.Intent {
	message = strings.ToLower(message)
	bestIntent := models.Intent{
		Name:       "general",
		Confidence: 0.0,
		Entities:   make(map[string]interface{}),
	}

	for _, pattern := range s.patterns {
		confidence := 0.0

		// Check patterns
		for _, p := range pattern.Patterns {
			if strings.Contains(message, strings.ToLower(p)) {
				confidence += 0.8
				break
			}
		}

		// Check keywords
		keywordMatches := 0
		for _, keyword := range pattern.Keywords {
			if strings.Contains(message, keyword) {
				keywordMatches++
			}
		}

		if len(pattern.Keywords) > 0 {
			confidence += float64(keywordMatches) / float64(len(pattern.Keywords)) * 0.4
		}

		if confidence > bestIntent.Confidence {
			bestIntent = models.Intent{
				Name:       pattern.Name,
				Confidence: confidence,
				Entities:   s.extractEntities(message, pattern.Name),
			}
		}
	}

	return bestIntent
}

func (s *IntentService) extractEntities(message string, intentName string) map[string]interface{} {
	entities := make(map[string]interface{})

	// Extract amounts
	amountRegex := regexp.MustCompile(`(?:₹|rs\.?|rupees?)?\s*(\d+(?:,\d{3})*(?:\.\d{2})?)\s*(?:rs|rupees|inr)?`)
	if matches := amountRegex.FindStringSubmatch(message); len(matches) > 1 {
		entities["amount"] = strings.ReplaceAll(matches[1], ",", "")
	}

	// Extract names (for payees)
	if strings.Contains(intentName, "payee") {
		nameRegex := regexp.MustCompile(`(?:payee|to|for)\s+([A-Za-z\s]+)`)
		if matches := nameRegex.FindStringSubmatch(message); len(matches) > 1 {
			entities["payee_name"] = strings.TrimSpace(matches[1])
		}
	}

	// Extract loan types
	if strings.Contains(intentName, "loan") {
		loanTypes := []string{"personal", "home", "car", "business"}
		for _, loanType := range loanTypes {
			if strings.Contains(message, loanType) {
				entities["loan_type"] = loanType
				break
			}
		}
	}

	// Extract transfer methods
	methods := []string{"upi", "imps", "neft", "rtgs"}
	for _, method := range methods {
		if strings.Contains(message, method) {
			entities["method"] = strings.ToUpper(method)
			break
		}
	}

	return entities
}

func (s *IntentService) ExtractParametersFromMessage(message string, intent models.Intent) map[string]interface{} {
	params := make(map[string]interface{})

	// Add entities from intent recognition
	for key, value := range intent.Entities {
		params[key] = value
	}

	// Extract additional parameters based on intent
	switch intent.Name {
	case "fund_transfer":
		if _, exists := params["method"]; !exists {
			if strings.Contains(strings.ToLower(message), "upi") {
				params["method"] = "UPI"
			} else if strings.Contains(strings.ToLower(message), "imps") {
				params["method"] = "IMPS"
			} else if strings.Contains(strings.ToLower(message), "neft") {
				params["method"] = "NEFT"
			} else if strings.Contains(strings.ToLower(message), "rtgs") {
				params["method"] = "RTGS"
			} else {
				params["method"] = "UPI" // Default
			}
		}
	case "loan_application":
		// Set default action if not specified
		if _, exists := params["action"]; !exists {
			if strings.Contains(strings.ToLower(message), "apply") {
				params["action"] = "apply"
			} else if strings.Contains(strings.ToLower(message), "eligibility") {
				params["action"] = "eligibility"
			} else if strings.Contains(strings.ToLower(message), "calculate") || strings.Contains(strings.ToLower(message), "emi") {
				params["action"] = "calculate"
			} else {
				params["action"] = "info"
			}
		}
	}

	return params
}

func (s *IntentService) initializeIntentPatterns() {
	s.patterns = []models.IntentPattern{
		{
			Name:     "fund_transfer",
			Patterns: []string{"transfer money", "send money", "pay", "transfer funds"},
			Keywords: []string{"transfer", "send", "pay", "money", "funds", "upi", "imps", "neft"},
			Action:   "fund_transfer",
		},
		{
			Name:     "check_balance",
			Patterns: []string{"check balance", "account balance", "show balance"},
			Keywords: []string{"balance", "account", "check", "show"},
			Action:   "view_balance",
		},
		{
			Name:     "add_payee",
			Patterns: []string{"add payee", "new payee", "add beneficiary"},
			Keywords: []string{"add", "payee", "beneficiary", "new", "register"},
			Action:   "add_payee",
		},
		{
			Name:     "loan_application",
			Patterns: []string{"apply loan", "loan application", "need loan"},
			Keywords: []string{"loan", "apply", "borrow", "credit", "emi"},
			Action:   "loan_application",
		},
	}
}
