package agents

import (
	"fmt"
	"strings"

	"github.com/banking/ai-agents-banking/src/dao"
	"github.com/banking/ai-agents-banking/src/models"
)

type AccountBalanceAgent struct {
	*BaseAgent
	accountDAO *dao.AccountDAO
}

func NewAccountBalanceAgent(accountDAO *dao.AccountDAO) *AccountBalanceAgent {
	return &AccountBalanceAgent{
		BaseAgent: &BaseAgent{
			Name:        "AccountBalanceAgent",
			Description: "Provides account balance and account information",
			Tools:       []string{"check_balance", "download_statement", "view_transactions", "set_alerts"},
			Confidence:  0.95,
		},
		accountDAO: accountDAO,
	}
}

func (a *AccountBalanceAgent) CanHandle(intent string, message string) bool {
	balanceIntents := []string{"check_balance", "view_balance", "account_balance", "balance_inquiry"}
	for _, bi := range balanceIntents {
		if intent == bi {
			return true
		}
	}

	balanceKeywords := []string{"balance", "account", "check", "show", "statement"}
	lowerMsg := strings.ToLower(message)
	for _, keyword := range balanceKeywords {
		if strings.Contains(lowerMsg, keyword) {
			return true
		}
	}
	return false
}

func (a *AccountBalanceAgent) GetRequiredParameters() []string {
	return []string{} // No required parameters
}

func (a *AccountBalanceAgent) ValidateParameters(params map[string]interface{}) []string {
	return []string{} // Always valid
}

func (a *AccountBalanceAgent) Process(ctx *models.AgentContext) *models.AgentResponse {
	accounts, err := a.accountDAO.GetUserAccounts(ctx.UserID)
	if err != nil {
		return &models.AgentResponse{
			Message:   "Failed to retrieve account information",
			AgentName: a.Name,
			Data:      err,
		}
	}

	if len(accounts) == 0 {
		return &models.AgentResponse{
			Message:   "No accounts found for this user",
			AgentName: a.Name,
			Data:      nil,
		}
	}

	var response strings.Builder
	response.WriteString("💳 **Your Account Information**\n\n")

	totalBalance := 0.0
	for i, account := range accounts {
		response.WriteString(fmt.Sprintf("**Account %d - %s**\n", i+1, account.AccountType))
		response.WriteString(fmt.Sprintf("• Account Number: ****%s\n", account.AccountNumber[len(account.AccountNumber)-4:]))
		response.WriteString(fmt.Sprintf("• Available Balance: ₹%.2f\n", account.Balance))
		response.WriteString(fmt.Sprintf("• Current Balance: ₹%.2f\n", account.Balance))
		response.WriteString(fmt.Sprintf("• Last Updated: %s\n\n", account.LastUpdated.Format("02 Jan 2006, 15:04")))
		totalBalance += account.Balance
	}

	response.WriteString(fmt.Sprintf("💰 **Total Balance: ₹%.2f**", totalBalance))

	return &models.AgentResponse{
		Message: response.String(),
		Data: map[string]interface{}{
			"accounts":      accounts,
			"total_balance": totalBalance,
		},
		Actions:   a.Tools,
		AgentName: a.Name,
	}
}

func (a *AccountBalanceAgent) GetHelp() string {
	return `💳 **Account Balance Agent Help**

I can help you check your account information:

**What I provide:**
• Current account balance
• Available balance
• Account details
• Multiple account summary

**Example commands:**
• "Check my balance"
• "Show account details"
• "What's my current balance?"
• "Account summary"`
}
