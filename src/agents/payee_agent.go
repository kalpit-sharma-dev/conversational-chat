package agents

import (
	"fmt"
	"strings"
	"time"

	"github.com/banking/ai-agents-banking/src/dao"
	"github.com/banking/ai-agents-banking/src/models"
	"github.com/banking/ai-agents-banking/src/utils"
)

type AddPayeeAgent struct {
	*BaseAgent
	payeeDAO *dao.PayeeDAO
}

func NewAddPayeeAgent(payeeDAO *dao.PayeeDAO) *AddPayeeAgent {
	return &AddPayeeAgent{
		BaseAgent: &BaseAgent{
			Name:        "AddPayeeAgent",
			Description: "Manages adding new payees and beneficiaries",
			Tools:       []string{"add_payee", "verify_payee", "transfer_to_payee", "view_all_payees"},
			Confidence:  0.9,
		},
		payeeDAO: payeeDAO,
	}
}

func (a *AddPayeeAgent) CanHandle(intent string, message string) bool {
	payeeIntents := []string{"add_payee", "add_beneficiary", "new_payee", "register_payee"}
	for _, pi := range payeeIntents {
		if intent == pi {
			return true
		}
	}

	payeeKeywords := []string{"add payee", "new payee", "register", "beneficiary", "add recipient"}
	lowerMsg := strings.ToLower(message)
	for _, keyword := range payeeKeywords {
		if strings.Contains(lowerMsg, keyword) {
			return true
		}
	}
	return false
}

func (a *AddPayeeAgent) GetRequiredParameters() []string {
	return []string{"payee_name", "account_number", "ifsc_code"}
}

func (a *AddPayeeAgent) ValidateParameters(params map[string]interface{}) []string {
	required := a.GetRequiredParameters()
	var missing []string

	for _, param := range required {
		if _, exists := params[param]; !exists {
			missing = append(missing, param)
		}
	}

	// Validate IFSC if present
	if ifsc, exists := params["ifsc_code"]; exists {
		if !utils.IsValidIFSC(fmt.Sprintf("%v", ifsc)) {
			missing = append(missing, "valid_ifsc")
		}
	}

	// Validate account number if present
	if accNo, exists := params["account_number"]; exists {
		if !utils.IsValidAccountNumber(fmt.Sprintf("%v", accNo)) {
			missing = append(missing, "valid_account")
		}
	}

	return missing
}

func (a *AddPayeeAgent) Process(ctx *models.AgentContext) *models.AgentResponse {
	missing := a.ValidateParameters(ctx.Parameters)
	if len(missing) > 0 {
		return &models.AgentResponse{
			Message:           a.getQuestionForMissing(missing[0]),
			AgentName:         a.Name,
			RequiresInput:     true,
			MissingParameters: missing,
		}
	}

	payeeName := fmt.Sprintf("%v", ctx.Parameters["payee_name"])
	accountNumber := fmt.Sprintf("%v", ctx.Parameters["account_number"])
	ifscCode := fmt.Sprintf("%v", ctx.Parameters["ifsc_code"])

	// Check if payee already exists
	existingPayee, err := a.payeeDAO.GetPayeeByAccount(ctx.UserID, accountNumber)
	if err != nil {
		return &models.AgentResponse{
			Message:   "Failed to check existing payees",
			AgentName: a.Name,
			Data:      err,
		}
	}

	if existingPayee != nil {
		return &models.AgentResponse{
			Message:   fmt.Sprintf("Payee with account number %s already exists as '%s'", accountNumber, existingPayee.Name),
			AgentName: a.Name,
			Data:      existingPayee,
		}
	}

	// Get bank name from IFSC
	bankName := utils.GetBankNameFromIFSC(ifscCode)

	// Create new payee
	payee := models.Payee{
		ID:         fmt.Sprintf("PAYEE_%d", time.Now().UnixNano()),
		Name:       payeeName,
		AccountNo:  accountNumber,
		BankName:   bankName,
		IFSCCode:   ifscCode,
		PayeeType:  "External",
		AddedDate:  time.Now(),
		IsActive:   true,
		IsVerified: false, // Would be verified via OTP in real system
	}

	// Add to user's payee list
	err = a.payeeDAO.AddUserPayee(ctx.UserID, payee)
	if err != nil {
		return &models.AgentResponse{
			Message:   "Failed to add payee",
			AgentName: a.Name,
			Data:      err,
		}
	}

	return &models.AgentResponse{
		Message: fmt.Sprintf("✅ **Payee Added Successfully!**\n\n👤 Name: %s\n🏦 Bank: %s\n💳 Account: ****%s\n🏛️ IFSC: %s\n\n⚠️ Payee will be verified within 24 hours for enhanced security.",
			payeeName, bankName, accountNumber[len(accountNumber)-4:], ifscCode),
		Data:      payee,
		Actions:   a.Tools,
		AgentName: a.Name,
	}
}

func (a *AddPayeeAgent) getQuestionForMissing(param string) string {
	switch param {
	case "payee_name":
		return "👤 What's the payee's name?"
	case "account_number":
		return "💳 Please provide the account number:"
	case "ifsc_code":
		return "🏛️ What's the IFSC code of the bank?"
	case "valid_ifsc":
		return "❌ Please provide a valid IFSC code (e.g., SBIN0001234)"
	case "valid_account":
		return "❌ Please provide a valid account number"
	default:
		return fmt.Sprintf("Please provide the %s", param)
	}
}

func (a *AddPayeeAgent) GetHelp() string {
	return `👤 **Add Payee Agent Help**

I can help you add new payees for transfers:

**What I need:**
• Payee name
• Account number
• IFSC code
• Bank details (optional)

**Security Features:**
• Duplicate detection
• IFSC validation
• 24-hour verification period

**Example commands:**
• "Add new payee John"
• "Register beneficiary for HDFC account"
• "Add payee with account 1234567890"`
}
