package dao

import (
	"fmt"
	"sync"
	"time"

	"github.com/banking/ai-agents-banking/src/models"
)

type AccountDAO struct {
	userAccounts map[string][]models.Account
	mu           sync.RWMutex
}

func NewAccountDAO() *AccountDAO {
	dao := &AccountDAO{
		userAccounts: make(map[string][]models.Account),
	}
	dao.initializeMockData()
	return dao
}

func (d *AccountDAO) GetUserAccounts(userID string) ([]models.Account, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if accounts, exists := d.userAccounts[userID]; exists {
		return accounts, nil
	}

	return []models.Account{}, nil
}

func (d *AccountDAO) GetUserAccount(userID string) (*models.Account, error) {
	accounts, err := d.GetUserAccounts(userID)
	if err != nil {
		return nil, err
	}
	if len(accounts) == 0 {
		return nil, fmt.Errorf("no accounts found for user")
	}
	return &accounts[0], nil
}

func (d *AccountDAO) GetAccountByID(userID, accountID string) (*models.Account, error) {
	accounts, err := d.GetUserAccounts(userID)
	if err != nil {
		return nil, err
	}

	for _, account := range accounts {
		if account.AccountID == accountID {
			return &account, nil
		}
	}

	return nil, fmt.Errorf("account not found")
}

func (d *AccountDAO) UpdateAccountBalance(userID, accountID string, newBalance float64) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	accounts, exists := d.userAccounts[userID]
	if !exists {
		return fmt.Errorf("no accounts found for user")
	}

	for i, account := range accounts {
		if account.AccountID == accountID {
			accounts[i].Balance = newBalance
			accounts[i].LastUpdated = time.Now()
			d.userAccounts[userID] = accounts
			return nil
		}
	}

	return fmt.Errorf("account not found")
}

func (d *AccountDAO) initializeMockData() {
	d.userAccounts["user123"] = []models.Account{
		{
			AccountID:      "ACC_001",
			UserID:         "user123",
			AccountNumber:  "1234567890",
			AccountType:    "Savings",
			Balance:        150000.00,
			Currency:       "INR",
			Status:         "Active",
			OpeningDate:    time.Now().AddDate(-1, 0, 0),
			LastUpdated:    time.Now(),
			BranchCode:     "001234",
			IFSCCode:       "BANK0001234",
			InterestRate:   3.5,
			MinimumBalance: 1000.00,
		},
		{
			AccountID:      "ACC_002",
			UserID:         "user123",
			AccountNumber:  "0987654321",
			AccountType:    "Current",
			Balance:        250000.00,
			Currency:       "INR",
			Status:         "Active",
			OpeningDate:    time.Now().AddDate(-2, 0, 0),
			LastUpdated:    time.Now(),
			BranchCode:     "001234",
			IFSCCode:       "BANK0001234",
			MinimumBalance: 5000.00,
		},
	}
}
