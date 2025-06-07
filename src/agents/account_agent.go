package agents

import (
	"context"
	"fmt"

	"github.com/banking/ai-agents-banking/src/dao"
	"github.com/banking/ai-agents-banking/src/models"
)

type AccountAgent struct {
	repo *dao.AccountDAO
}

// GetAccount retrieves an account by its ID
func (a *AccountAgent) GetAccount(ctx context.Context, userID, accountID string) (*models.Account, error) {
	account, err := a.repo.GetAccountByID(userID, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}
	return account, nil
}

// GetAccountByID retrieves an account by its ID
func (a *AccountAgent) GetAccountByID(ctx context.Context, userID, accountID string) (*models.Account, error) {
	account, err := a.repo.GetAccountByID(userID, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get account by ID: %w", err)
	}
	return account, nil
}

// GetAccountByNumber retrieves an account by its account number
func (a *AccountAgent) GetAccountByNumber(ctx context.Context, userID, accountNumber string) (*models.Account, error) {
	accounts, err := a.repo.GetUserAccounts(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts: %w", err)
	}
	for _, account := range accounts {
		if account.AccountNumber == accountNumber {
			return &account, nil
		}
	}
	return nil, fmt.Errorf("account not found")
}
