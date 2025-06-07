package services

import (
	"fmt"
	"time"

	"github.com/banking/ai-agents-banking/src/agents"
	"github.com/banking/ai-agents-banking/src/dao"
	"github.com/banking/ai-agents-banking/src/models"
)

type AgentService struct {
	agents      map[string]agents.BankingAgent
	fallback    agents.BankingAgent
	accountDAO  *dao.AccountDAO
	payeeDAO    *dao.PayeeDAO
	transferDAO *dao.TransferDAO
	loanDAO     *dao.LoanDAO
}

func NewAgentService(accountDAO *dao.AccountDAO, payeeDAO *dao.PayeeDAO, transferDAO *dao.TransferDAO, loanDAO *dao.LoanDAO) *AgentService {
	service := &AgentService{
		agents:      make(map[string]agents.BankingAgent),
		accountDAO:  accountDAO,
		payeeDAO:    payeeDAO,
		transferDAO: transferDAO,
		loanDAO:     loanDAO,
	}

	// Register all agents
	service.RegisterAgent(agents.NewFundTransferAgent(accountDAO, payeeDAO, transferDAO))
	service.RegisterAgent(agents.NewAccountBalanceAgent(accountDAO))
	service.RegisterAgent(agents.NewAddPayeeAgent(payeeDAO))
	service.RegisterAgent(agents.NewLoanAgent(loanDAO))

	// Set fallback agent
	service.fallback = agents.NewGeneralBankingAgent()

	return service
}

func (s *AgentService) RegisterAgent(agent agents.BankingAgent) {
	s.agents[agent.GetName()] = agent
}

func (s *AgentService) GetAgent(intent string, message string) agents.BankingAgent {
	// Find the best matching agent
	for _, agent := range s.agents {
		if agent.CanHandle(intent, message) {
			return agent
		}
	}

	// Return fallback if no specific agent can handle
	return s.fallback
}

func (s *AgentService) ProcessWithAgent(ctx *models.AgentContext) *models.AgentResponse {
	agent := s.GetAgent(ctx.Intent, ctx.Message)
	return agent.Process(ctx)
}

func (s *AgentService) GetAllAgents() map[string]agents.BankingAgent {
	result := make(map[string]agents.BankingAgent)
	for name, agent := range s.agents {
		result[name] = agent
	}
	return result
}

func (s *AgentService) GetAgentsCount() int {
	return len(s.agents)
}

// Banking operations
func (s *AgentService) ExecuteTransfer(amount float64, recipient string, method string) (interface{}, error) {
	// Create transfer request
	transfer := models.TransferRequest{
		TransferID: fmt.Sprintf("TXN%d", time.Now().UnixNano()),
		Amount:     amount,
		Method:     method,
		Status:     "SUCCESS",
		Timestamp:  time.Now(),
		Reference:  fmt.Sprintf("REF%d", time.Now().UnixNano()),
		Fees:       0, // Calculate based on method
	}

	// Store transfer
	transferRecord := models.Transfer{
		TransferID:    transfer.TransferID,
		FromAccountID: "", // Set appropriately if you have account info
		ToAccountID:   "", // Set appropriately if you have recipient info
		Amount:        transfer.Amount,
		Method:        transfer.Method,
		Status:        transfer.Status,
		Timestamp:     transfer.Timestamp,
		Reference:     transfer.Reference,
		Fees:          transfer.Fees,
		Description:   transfer.Description,
	}
	err := s.transferDAO.AddTransfer("user_id", transferRecord)
	if err != nil {
		return nil, err
	}

	return transfer, nil
}

func (s *AgentService) GetBalance(accountID string) (float64, error) {
	account, err := s.accountDAO.GetUserAccount(accountID)
	if err != nil {
		return 0, err
	}
	return account.Balance, nil
}

func (s *AgentService) AddPayee(name string, accountNumber string) (*models.Payee, error) {
	payee := &models.Payee{
		ID:         fmt.Sprintf("PAYEE_%d", time.Now().UnixNano()),
		Name:       name,
		AccountNo:  accountNumber,
		PayeeType:  "External",
		AddedDate:  time.Now(),
		IsActive:   true,
		IsVerified: false,
	}

	err := s.payeeDAO.AddUserPayee("user_id", *payee)
	if err != nil {
		return nil, err
	}

	return payee, nil
}

func (s *AgentService) CreateFixedDeposit(amount float64, tenure int, tenureUnit string) (interface{}, error) {
	// Mock FD creation
	fd := map[string]interface{}{
		"id":            fmt.Sprintf("FD_%d", time.Now().UnixNano()),
		"amount":        amount,
		"tenure":        tenure,
		"tenure_unit":   tenureUnit,
		"created_at":    time.Now(),
		"maturity_at":   time.Now().AddDate(0, tenure, 0),
		"interest_rate": 6.5, // Mock rate
	}

	return fd, nil
}

func (s *AgentService) CreateRecurringDeposit(amount float64, tenure int, tenureUnit string) (interface{}, error) {
	// Mock RD creation
	rd := map[string]interface{}{
		"id":            fmt.Sprintf("RD_%d", time.Now().UnixNano()),
		"amount":        amount,
		"tenure":        tenure,
		"tenure_unit":   tenureUnit,
		"created_at":    time.Now(),
		"maturity_at":   time.Now().AddDate(0, tenure, 0),
		"interest_rate": 7.0, // Mock rate
	}

	return rd, nil
}

func (s *AgentService) GetInterestRates(productType string) (interface{}, error) {
	// Mock interest rates
	rates := map[string]interface{}{
		"personal_loan": 12.5,
		"home_loan":     8.5,
		"car_loan":      9.5,
		"fd":            6.5,
		"rd":            7.0,
	}

	if productType != "all" {
		if rate, exists := rates[productType]; exists {
			return map[string]interface{}{productType: rate}, nil
		}
		return nil, fmt.Errorf("invalid product type")
	}

	return rates, nil
}
