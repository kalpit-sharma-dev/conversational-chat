package services

import (
	"fmt"
	"time"
)

// Base banking tool
type BaseBankingTool struct {
	AgentService *AgentService
}

// Fund Transfer Tool
type FundTransferTool struct {
	BaseBankingTool
}

func (t *FundTransferTool) Name() string {
	return "fund_transfer"
}

func (t *FundTransferTool) Description() string {
	return "Transfer funds between accounts using UPI/IMPS/NEFT"
}

func (t *FundTransferTool) Execute(params map[string]interface{}) (interface{}, error) {
	amount, ok := params["amount"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid amount")
	}

	recipient, ok := params["recipient"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid recipient")
	}

	method, ok := params["method"].(string)
	if !ok {
		method = "UPI" // Default to UPI
	}

	// Execute transfer through agent service
	result, err := t.AgentService.ExecuteTransfer(amount, recipient, method)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Balance Check Tool
type BalanceCheckTool struct {
	BaseBankingTool
}

func (t *BalanceCheckTool) Name() string {
	return "check_balance"
}

func (t *BalanceCheckTool) Description() string {
	return "Check account balance"
}

func (t *BalanceCheckTool) Execute(params map[string]interface{}) (interface{}, error) {
	accountID, ok := params["account_id"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid account ID")
	}

	// Get balance through agent service
	balance, err := t.AgentService.GetBalance(accountID)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"account_id": accountID,
		"balance":    balance,
	}, nil
}

// Add Payee Tool
type AddPayeeTool struct {
	BaseBankingTool
}

func (t *AddPayeeTool) Name() string {
	return "add_payee"
}

func (t *AddPayeeTool) Description() string {
	return "Add a new payee/beneficiary"
}

func (t *AddPayeeTool) Execute(params map[string]interface{}) (interface{}, error) {
	name, ok := params["name"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid payee name")
	}

	accountNumber, ok := params["account_number"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid account number")
	}

	// Add payee through agent service
	payee, err := t.AgentService.AddPayee(name, accountNumber)
	if err != nil {
		return nil, err
	}

	return payee, nil
}

// Fixed Deposit Tool
type FixedDepositTool struct {
	BaseBankingTool
}

func (t *FixedDepositTool) Name() string {
	return "create_fd"
}

func (t *FixedDepositTool) Description() string {
	return "Create a fixed deposit"
}

func (t *FixedDepositTool) Execute(params map[string]interface{}) (interface{}, error) {
	amount, ok := params["amount"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid amount")
	}

	tenure, ok := params["tenure"].(int)
	if !ok {
		return nil, fmt.Errorf("invalid tenure")
	}

	tenureUnit, ok := params["tenure_unit"].(string)
	if !ok {
		tenureUnit = "months" // Default to months
	}

	// Create FD through agent service
	fd, err := t.AgentService.CreateFixedDeposit(amount, tenure, tenureUnit)
	if err != nil {
		return nil, err
	}

	return fd, nil
}

// Recurring Deposit Tool
type RecurringDepositTool struct {
	BaseBankingTool
}

func (t *RecurringDepositTool) Name() string {
	return "create_rd"
}

func (t *RecurringDepositTool) Description() string {
	return "Create a recurring deposit"
}

func (t *RecurringDepositTool) Execute(params map[string]interface{}) (interface{}, error) {
	amount, ok := params["amount"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid amount")
	}

	tenure, ok := params["tenure"].(int)
	if !ok {
		return nil, fmt.Errorf("invalid tenure")
	}

	tenureUnit, ok := params["tenure_unit"].(string)
	if !ok {
		tenureUnit = "months" // Default to months
	}

	// Create RD through agent service
	rd, err := t.AgentService.CreateRecurringDeposit(amount, tenure, tenureUnit)
	if err != nil {
		return nil, err
	}

	return rd, nil
}

// Interest Rates Tool
type InterestRatesTool struct {
	BaseBankingTool
}

func (t *InterestRatesTool) Name() string {
	return "get_interest_rates"
}

func (t *InterestRatesTool) Description() string {
	return "Get current interest rates for various products"
}

func (t *InterestRatesTool) Execute(params map[string]interface{}) (interface{}, error) {
	productType, ok := params["product_type"].(string)
	if !ok {
		productType = "all" // Default to all products
	}

	// Get rates through agent service
	rates, err := t.AgentService.GetInterestRates(productType)
	if err != nil {
		return nil, err
	}

	return rates, nil
}

// Weather Tool
type WeatherTool struct {
	CacheTTL time.Duration
	Cache    map[string]interface{}
}

func (t *WeatherTool) Name() string {
	return "get_weather"
}

func (t *WeatherTool) Description() string {
	return "Get weather information for a location"
}

func (t *WeatherTool) Execute(params map[string]interface{}) (interface{}, error) {
	location, ok := params["location"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid location")
	}

	// Check cache first
	if cached, exists := t.Cache[location]; exists {
		return cached, nil
	}

	// Simulate weather API call
	weather := map[string]interface{}{
		"location":  location,
		"temp":      25.5,
		"condition": "Sunny",
		"humidity":  65,
	}

	// Cache the result
	t.Cache[location] = weather

	return weather, nil
}
