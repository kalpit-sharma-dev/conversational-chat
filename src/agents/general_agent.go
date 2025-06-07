package agents

import (
	"strings"

	"github.com/banking/ai-agents-banking/src/models"
)

type GeneralBankingAgent struct {
	*BaseAgent
}

func NewGeneralBankingAgent() *GeneralBankingAgent {
	return &GeneralBankingAgent{
		BaseAgent: &BaseAgent{
			Name:        "GeneralBankingAgent",
			Description: "Handles general banking queries and provides overall banking assistance",
			Tools:       []string{"transfer_money", "check_balance", "add_payee", "apply_loan"},
			Confidence:  0.8,
		},
	}
}

func (a *GeneralBankingAgent) CanHandle(intent string, message string) bool {
	return true // Fallback agent handles everything
}

func (a *GeneralBankingAgent) Process(ctx *models.AgentContext) *models.AgentResponse {
	// This agent provides general responses and guidance
	response := a.generateGeneralResponse(ctx.Message)

	return &models.AgentResponse{
		Message:       response,
		AgentName:     a.Name,
		Actions:       a.Tools,
		RequiresInput: false,
		RequiresTool:  false,
		Data:          nil,
	}
}

func (a *GeneralBankingAgent) generateGeneralResponse(message string) string {
	lowerMsg := strings.ToLower(message)

	if strings.Contains(lowerMsg, "help") || strings.Contains(lowerMsg, "what can you do") {
		return `🏦 **Welcome to AI Banking Assistant!**

I can help you with:

**🔄 Fund Transfers**
• UPI, IMPS, NEFT, RTGS transfers
• Quick payments to saved payees

**💳 Account Services**
• Check account balance
• View transaction history
• Account statements

**👥 Payee Management**
• Add new beneficiaries
• Manage saved payees
• Verify payee details

**🏦 Loan Services**
• Check loan eligibility
• Apply for loans
• Calculate EMI
• Track applications

**Example commands:**
• "Transfer ₹5000 to John"
• "Check my balance"
• "Add new payee"
• "Apply for personal loan"`
	}

	if strings.Contains(lowerMsg, "thank") {
		return "You're welcome! I'm here to help with all your banking needs. Is there anything else you'd like to do?"
	}

	if strings.Contains(lowerMsg, "goodbye") || strings.Contains(lowerMsg, "bye") {
		return "Thank you for using our banking services! Have a great day! 🙏"
	}

	// Default general response
	return `🏦 I'm your AI banking assistant, ready to help you with various banking services.

**Quick Actions:**
• Transfer money
• Check balance
• Add payee
• Apply for loans

Please let me know what you'd like to do, or say "help" to see all available services.`
}

func (a *GeneralBankingAgent) GetHelp() string {
	return `🏦 **General Banking Agent Help**

I'm your general banking assistant and can help with:

**Available Services:**
• Fund transfers and payments
• Account balance inquiries
• Payee management
• Loan applications and information
• General banking queries

**Specialized Agents:**
• Fund Transfer Agent - Money transfers
• Account Balance Agent - Balance inquiries
• Add Payee Agent - Beneficiary management
• Loan Agent - Loan services

**Example commands:**
• "What can you help me with?"
• "Show me my options"
• "Help with banking services"`
}
