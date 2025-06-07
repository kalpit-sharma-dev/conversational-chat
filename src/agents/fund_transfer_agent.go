package agents

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/banking/ai-agents-banking/src/dao"
	"github.com/banking/ai-agents-banking/src/models"
	"github.com/banking/ai-agents-banking/src/utils"
)

type FundTransferAgent struct {
	*BaseAgent
	accountDAO  *dao.AccountDAO
	payeeDAO    *dao.PayeeDAO
	transferDAO *dao.TransferDAO
}

func NewFundTransferAgent(accountDAO *dao.AccountDAO, payeeDAO *dao.PayeeDAO, transferDAO *dao.TransferDAO) *FundTransferAgent {
	return &FundTransferAgent{
		BaseAgent: &BaseAgent{
			Name:        "FundTransferAgent",
			Description: "Handles money transfers via UPI, IMPS, NEFT, and RTGS",
			Tools:       []string{"transfer_money", "view_transaction_details", "download_receipt"},
			Confidence:  0.9,
		},
		accountDAO:  accountDAO,
		payeeDAO:    payeeDAO,
		transferDAO: transferDAO,
	}
}

func (a *FundTransferAgent) CanHandle(intent string, message string) bool {
	transferIntents := []string{"fund_transfer", "send_money", "transfer_money", "pay_money", "upi_transfer"}
	for _, ti := range transferIntents {
		if intent == ti {
			return true
		}
	}

	transferKeywords := []string{"transfer", "send", "pay", "money", "upi", "imps", "neft", "rtgs"}
	lowerMsg := strings.ToLower(message)
	for _, keyword := range transferKeywords {
		if strings.Contains(lowerMsg, keyword) {
			return true
		}
	}
	return false
}

func (a *FundTransferAgent) GetRequiredParameters() []string {
	return []string{"amount", "method"}
}

func (a *FundTransferAgent) ValidateParameters(params map[string]interface{}) []string {
	required := a.GetRequiredParameters()
	var missing []string

	for _, param := range required {
		if _, exists := params[param]; !exists {
			missing = append(missing, param)
		}
	}

	// Validate amount if present
	if amountStr, exists := params["amount"]; exists {
		if amount, err := strconv.ParseFloat(fmt.Sprintf("%v", amountStr), 64); err != nil || amount <= 0 {
			missing = append(missing, "valid_amount")
		}
	}

	return missing
}

func (a *FundTransferAgent) Process(ctx *models.AgentContext) *models.AgentResponse {
	missing := a.ValidateParameters(ctx.Parameters)
	if len(missing) > 0 {
		return &models.AgentResponse{
			Message:           a.getQuestionForMissing(missing[0]),
			AgentName:         a.Name,
			RequiresInput:     true,
			MissingParameters: missing,
		}
	}

	amount, _ := strconv.ParseFloat(fmt.Sprintf("%v", ctx.Parameters["amount"]), 64)
	method := fmt.Sprintf("%v", ctx.Parameters["method"])

	// Check user's account balance
	userAccount, err := a.accountDAO.GetUserAccount(ctx.UserID)
	if err != nil {
		return &models.AgentResponse{
			Message:   "Failed to retrieve account information",
			AgentName: a.Name,
			Data:      err,
		}
	}

	if userAccount.Balance < amount {
		return &models.AgentResponse{
			Message:   fmt.Sprintf("Insufficient balance. Available: ₹%.2f, Required: ₹%.2f", userAccount.Balance, amount),
			AgentName: a.Name,
			Data:      map[string]interface{}{"available_balance": userAccount.Balance, "required_amount": amount},
		}
	}

	// Process transfer
	transferID := fmt.Sprintf("TXN%d", time.Now().UnixNano())
	fees := utils.CalculateTransferFees(method, amount)

	transfer := models.TransferRequest{
		TransferID: transferID,
		Amount:     amount,
		Method:     method,
		Status:     "SUCCESS",
		Timestamp:  time.Now(),
		Reference:  fmt.Sprintf("REF%d", time.Now().UnixNano()),
		Fees:       fees,
	}

	// Update account balance
	a.accountDAO.UpdateAccountBalance(ctx.UserID, userAccount.AccountNumber, userAccount.Balance-amount-fees)

	// Store transfer history
	transferRecord := models.Transfer{
		TransferID:    transfer.TransferID,
		FromAccountID: userAccount.AccountID,
		ToAccountID:   "", // Set appropriately if you have payee info
		Amount:        transfer.Amount,
		Method:        transfer.Method,
		Status:        transfer.Status,
		Timestamp:     transfer.Timestamp,
		Reference:     transfer.Reference,
		Fees:          transfer.Fees,
		Description:   transfer.Description,
	}
	a.transferDAO.AddTransfer(ctx.UserID, transferRecord)

	return &models.AgentResponse{
		Message: fmt.Sprintf("✅ Transfer completed successfully!\n💰 Amount: ₹%.2f\n🏦 Method: %s\n📋 Reference: %s\n💳 Transaction ID: %s\n💵 Fees: ₹%.2f",
			amount, method, transfer.Reference, transferID, fees),
		Data:      transfer,
		Actions:   a.Tools,
		AgentName: a.Name,
	}
}

func (a *FundTransferAgent) getQuestionForMissing(param string) string {
	switch param {
	case "amount":
		return "💰 How much would you like to transfer?"
	case "method":
		return "🏦 Which transfer method would you prefer?\n1. UPI (Instant)\n2. IMPS (Instant)\n3. NEFT (Up to 2 hours)\n4. RTGS (Real-time for large amounts)"
	case "payee", "to_account":
		return "👤 Who would you like to transfer money to? (Payee name or account number)"
	case "valid_amount":
		return "❌ Please enter a valid amount greater than 0"
	default:
		return fmt.Sprintf("Please provide the %s for the transfer", param)
	}
}

func (a *FundTransferAgent) GetHelp() string {
	return `🏦 **Fund Transfer Agent Help**

I can help you transfer money using various methods:

**Available Methods:**
• UPI - Instant transfers (₹1 to ₹1,00,000)
• IMPS - Instant transfers (24/7)
• NEFT - Batch processing (₹1 to ₹10,00,000)
• RTGS - Real-time (₹2,00,000+)

**What I need:**
• Transfer amount
• Transfer method
• Recipient details

**Example commands:**
• "Transfer ₹5000 via UPI"
• "Send money to John using IMPS"
• "Pay ₹15000 through NEFT"`
}
