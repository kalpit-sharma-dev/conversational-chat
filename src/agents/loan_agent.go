package agents

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/banking/ai-agents-banking/src/dao"
	"github.com/banking/ai-agents-banking/src/models"
)

type LoanAgent struct {
	*BaseAgent
	loanDAO *dao.LoanDAO
}

func NewLoanAgent(loanDAO *dao.LoanDAO) *LoanAgent {
	return &LoanAgent{
		BaseAgent: &BaseAgent{
			Name:        "LoanAgent",
			Description: "Handles loan applications, eligibility checks, and loan information",
			Tools:       []string{"apply_loan", "check_eligibility", "calculate_emi", "upload_documents", "check_status"},
			Confidence:  0.9,
		},
		loanDAO: loanDAO,
	}
}

func (a *LoanAgent) CanHandle(intent string, message string) bool {
	loanIntents := []string{"loan_application", "apply_loan", "loan_eligibility", "loan_info", "personal_loan", "home_loan"}
	for _, li := range loanIntents {
		if intent == li {
			return true
		}
	}

	loanKeywords := []string{"loan", "apply", "eligibility", "emi", "interest", "personal loan", "home loan", "car loan"}
	lowerMsg := strings.ToLower(message)
	for _, keyword := range loanKeywords {
		if strings.Contains(lowerMsg, keyword) {
			return true
		}
	}
	return false
}

func (a *LoanAgent) GetRequiredParameters() []string {
	return []string{"loan_type", "amount"}
}

func (a *LoanAgent) ValidateParameters(params map[string]interface{}) []string {
	var missing []string

	// Check loan type
	if _, exists := params["loan_type"]; !exists {
		missing = append(missing, "loan_type")
	}

	// Check amount for application
	if action, exists := params["action"]; exists && action == "apply" {
		if _, exists := params["amount"]; !exists {
			missing = append(missing, "amount")
		}
	}

	return missing
}

func (a *LoanAgent) Process(ctx *models.AgentContext) *models.AgentResponse {
	action := ctx.Parameters["action"]
	if action == nil {
		action = "info" // Default to showing info
	}

	switch action {
	case "apply":
		return a.handleLoanApplication(ctx)
	case "eligibility":
		return a.handleEligibilityCheck(ctx)
	case "calculate":
		return a.handleEMICalculation(ctx)
	default:
		return a.handleLoanInfo(ctx)
	}
}

func (a *LoanAgent) handleLoanInfo(ctx *models.AgentContext) *models.AgentResponse {
	loanProducts, err := a.loanDAO.GetLoanProducts()
	if err != nil {
		return &models.AgentResponse{
			Message:   "Failed to retrieve loan products",
			AgentName: a.Name,
			Data:      err,
		}
	}

	var response strings.Builder
	response.WriteString("🏦 **Available Loan Products**\n\n")

	for _, loan := range loanProducts {
		response.WriteString(fmt.Sprintf("**%s**\n", loan.Name))
		response.WriteString(fmt.Sprintf("• Interest Rate: %.2f%% per annum\n", loan.InterestRate))
		response.WriteString(fmt.Sprintf("• Amount Range: ₹%.0f - ₹%.0f\n", loan.MinAmount, loan.MaxAmount))
		response.WriteString(fmt.Sprintf("• Max Tenure: %d months\n", loan.MaxTenure))
		response.WriteString(fmt.Sprintf("• Processing Fee: %.2f%%\n\n", loan.ProcessingFee))
	}

	return &models.AgentResponse{
		Message:   response.String(),
		Data:      loanProducts,
		Actions:   a.Tools,
		AgentName: a.Name,
	}
}

func (a *LoanAgent) handleLoanApplication(ctx *models.AgentContext) *models.AgentResponse {
	missing := a.ValidateParameters(ctx.Parameters)
	if len(missing) > 0 {
		return &models.AgentResponse{
			Message:           a.getQuestionForMissing(missing[0]),
			AgentName:         a.Name,
			RequiresInput:     true,
			MissingParameters: missing,
		}
	}

	loanType := fmt.Sprintf("%v", ctx.Parameters["loan_type"])
	amount, _ := strconv.ParseFloat(fmt.Sprintf("%v", ctx.Parameters["amount"]), 64)

	// Get loan product details
	product, err := a.loanDAO.GetLoanProduct(loanType)
	if err != nil || product == nil {
		return &models.AgentResponse{
			Message:   "Invalid loan type or product not available",
			AgentName: a.Name,
			Data:      err,
		}
	}

	// Validate amount range
	if amount < product.MinAmount || amount > product.MaxAmount {
		return &models.AgentResponse{
			Message:   fmt.Sprintf("Amount should be between ₹%.0f and ₹%.0f for %s", product.MinAmount, product.MaxAmount, product.Name),
			AgentName: a.Name,
			Data:      map[string]interface{}{"min_amount": product.MinAmount, "max_amount": product.MaxAmount, "product": product.Name},
		}
	}

	// Create application
	application := models.LoanApplication{
		ApplicationID: fmt.Sprintf("LOAN_%d", time.Now().UnixNano()),
		LoanType:      loanType,
		Amount:        amount,
		Status:        "SUBMITTED",
		AppliedDate:   time.Now(),
		Documents:     []string{"income_proof", "address_proof", "bank_statements"},
	}

	// Calculate EMI
	tenure := 24 // Default tenure
	if t, exists := ctx.Parameters["tenure"]; exists {
		tenure, _ = strconv.Atoi(fmt.Sprintf("%v", t))
	}
	application.Tenure = tenure
	application.EMI = a.calculateEMI(amount, product.InterestRate, float64(tenure))

	return &models.AgentResponse{
		Message: fmt.Sprintf("✅ **Loan Application Submitted**\n\n📋 Application ID: %s\n💰 Amount: ₹%.2f\n🏦 Loan Type: %s\n💳 Estimated EMI: ₹%.2f\n\n📄 Required documents will be collected in the next step.",
			application.ApplicationID, amount, loanType, application.EMI),
		Data:      application,
		Actions:   a.Tools,
		AgentName: a.Name,
	}
}

func (a *LoanAgent) handleEligibilityCheck(ctx *models.AgentContext) *models.AgentResponse {
	// Mock eligibility check - in real system, this would check credit score, income, etc.
	eligibilityScore := 75.0       // Mock score
	maxEligibleAmount := 1000000.0 // Mock amount

	var response strings.Builder
	response.WriteString("📊 **Loan Eligibility Report**\n\n")
	response.WriteString(fmt.Sprintf("✅ Eligibility Score: %.0f/100\n", eligibilityScore))
	response.WriteString(fmt.Sprintf("💰 Max Eligible Amount: ₹%.2f\n", maxEligibleAmount))
	response.WriteString("📋 Status: **ELIGIBLE**\n\n")
	response.WriteString("**Eligible Products:**\n")
	response.WriteString("• Personal Loan (up to ₹10,00,000)\n")
	response.WriteString("• Car Loan (up to ₹25,00,000)\n")
	response.WriteString("• Home Loan (up to ₹1,00,00,000)")

	return &models.AgentResponse{
		Message: response.String(),
		Data: map[string]interface{}{
			"eligibility_score":   eligibilityScore,
			"max_eligible_amount": maxEligibleAmount,
			"status":              "ELIGIBLE",
		},
		Actions:   a.Tools,
		AgentName: a.Name,
	}
}

func (a *LoanAgent) handleEMICalculation(ctx *models.AgentContext) *models.AgentResponse {
	amount, _ := strconv.ParseFloat(fmt.Sprintf("%v", ctx.Parameters["amount"]), 64)
	rate, _ := strconv.ParseFloat(fmt.Sprintf("%v", ctx.Parameters["interest_rate"]), 64)
	tenure, _ := strconv.ParseFloat(fmt.Sprintf("%v", ctx.Parameters["tenure"]), 64)

	emi := a.calculateEMI(amount, rate, tenure)
	totalAmount := emi * tenure
	totalInterest := totalAmount - amount

	var response strings.Builder
	response.WriteString("💳 **EMI Calculation**\n\n")
	response.WriteString(fmt.Sprintf("💰 Loan Amount: ₹%.2f\n", amount))
	response.WriteString(fmt.Sprintf("📊 Interest Rate: %.2f%% p.a.\n", rate))
	response.WriteString(fmt.Sprintf("⏱️ Tenure: %.0f months\n", tenure))
	response.WriteString(fmt.Sprintf("💵 Monthly EMI: ₹%.2f\n", emi))
	response.WriteString(fmt.Sprintf("📈 Total Interest: ₹%.2f\n", totalInterest))
	response.WriteString(fmt.Sprintf("📊 Total Amount: ₹%.2f", totalAmount))

	return &models.AgentResponse{
		Message: response.String(),
		Data: map[string]interface{}{
			"loan_amount":    amount,
			"interest_rate":  rate,
			"tenure":         tenure,
			"emi":            emi,
			"total_interest": totalInterest,
			"total_amount":   totalAmount,
		},
		Actions:   a.Tools,
		AgentName: a.Name,
	}
}

func (a *LoanAgent) calculateEMI(principal, rate, tenure float64) float64 {
	// Convert annual rate to monthly and percentage to decimal
	monthlyRate := (rate / 12) / 100
	// EMI = P * r * (1 + r)^n / ((1 + r)^n - 1)
	emi := principal * monthlyRate * math.Pow(1+monthlyRate, tenure) / (math.Pow(1+monthlyRate, tenure) - 1)
	return math.Round(emi*100) / 100 // Round to 2 decimal places
}

func (a *LoanAgent) getQuestionForMissing(param string) string {
	switch param {
	case "loan_type":
		return "🏦 Which type of loan would you like to apply for?\n1. Personal Loan\n2. Home Loan\n3. Car Loan\n4. Education Loan"
	case "amount":
		return "💰 What loan amount would you like to apply for?"
	case "tenure":
		return "⏱️ What loan tenure (in months) would you prefer?"
	case "interest_rate":
		return "📊 What's the interest rate for this loan?"
	default:
		return fmt.Sprintf("Please provide the %s", param)
	}
}

func (a *LoanAgent) GetHelp() string {
	return `🏦 **Loan Agent Help**

I can help you with loan-related services:

**Available Services:**
• Loan applications
• Eligibility checks
• EMI calculations
• Loan information

**Loan Types:**
• Personal Loans
• Home Loans
• Car Loans
• Education Loans

**Example commands:**
• "Apply for personal loan"
• "Check my loan eligibility"
• "Calculate EMI for ₹500000"
• "Show available loan products"`
}
